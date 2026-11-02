package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrmotologa/indexlens/internal/explain"
	"github.com/alexandrmotologa/indexlens/internal/model"
	"github.com/alexandrmotologa/indexlens/internal/poller"
	"github.com/alexandrmotologa/indexlens/internal/synthesizer"
	"github.com/alexandrmotologa/indexlens/internal/watchdog"
)

// Server coordinates the HTTP REST API and telemetry providers.
type Server struct {
	client   *poller.Client
	demoMode bool
	mux      *http.ServeMux
}

// NewServer constructs the REST API server with live or demo fallback.
func NewServer(client *poller.Client, demoMode bool) *Server {
	s := &Server{
		client:   client,
		demoMode: demoMode || client == nil,
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// Handler returns the HTTP handler with CORS and security headers.
func (s *Server) Handler() http.Handler {
	return withMiddleware(s.mux)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/overview", s.handleOverview)
	s.mux.HandleFunc("GET /api/queries", s.handleQueries)
	s.mux.HandleFunc("GET /api/tables", s.handleTables)
	s.mux.HandleFunc("GET /api/cache", s.handleCache)
	s.mux.HandleFunc("POST /api/explain", s.handleExplain)
	s.mux.HandleFunc("GET /api/indexes/advisor", s.handleIndexesAdvisor)
	s.mux.HandleFunc("GET /api/autovacuum", s.handleAutovacuum)
	s.mux.HandleFunc("GET /api/bloat", s.handleBloat)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{
		"status":    "healthy",
		"demoMode":  s.demoMode,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	if s.demoMode {
		writeJSON(w, http.StatusOK, model.GetDemoOverview())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	overview, err := s.client.GetOverview(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to collect system overview: "+err.Error())
		return
	}

	// Enrich overview with cache and wraparound indicators
	cacheStats, err := s.client.GetCacheStats(ctx)
	if err == nil {
		overview.BufferCacheHitRatio = cacheStats.OverallHitRatio
	}

	wrapInfo, err := watchdog.CheckWraparound(ctx, s.client)
	if err == nil {
		overview.WraparoundSafetyPercent = 100.0 - wrapInfo.PercentConsumed
	}

	tables, err := s.client.GetTableStats(ctx)
	if err == nil {
		underIndexed := 0
		bloatWarnings := 0
		for _, t := range tables {
			totalScans := t.SeqScan + t.IdxScan
			if t.SeqScan >= 25 && (float64(t.SeqScan)/float64(totalScans+1) > 0.3 || t.SeqTupRead > 100000) {
				underIndexed++
			}
			if t.DeadTupleRatio > 20.0 {
				bloatWarnings++
			}
		}
		overview.UnderIndexedTablesCount = underIndexed
		overview.BloatWarningCount = bloatWarnings
	}

	writeJSON(w, http.StatusOK, overview)
}

func (s *Server) handleQueries(w http.ResponseWriter, r *http.Request) {
	if s.demoMode {
		writeJSON(w, http.StatusOK, model.GetDemoQueries())
		return
	}

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	sortBy := r.URL.Query().Get("sort")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	queries, err := s.client.GetSlowQueries(ctx, limit, sortBy)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read slow queries: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, queries)
}

func (s *Server) handleTables(w http.ResponseWriter, r *http.Request) {
	if s.demoMode {
		writeJSON(w, http.StatusOK, model.GetDemoTables())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tables, err := s.client.GetTableStats(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query table statistics: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, tables)
}

func (s *Server) handleCache(w http.ResponseWriter, r *http.Request) {
	if s.demoMode {
		writeJSON(w, http.StatusOK, model.GetDemoCacheStats())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	stats, err := s.client.GetCacheStats(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to calculate cache stats: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

type explainRequest struct {
	Plan  string `json:"plan"`
	Query string `json:"query"`
}

func (s *Server) handleExplain(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var req explainRequest
	_ = json.Unmarshal(body, &req)

	// If explicit JSON plan provided
	if strings.TrimSpace(req.Plan) != "" {
		result, err := explain.ParsePlanJSON([]byte(req.Plan))
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to parse EXPLAIN plan: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}

	// If SQL query provided and connected to PostgreSQL
	if strings.TrimSpace(req.Query) != "" && !s.demoMode && s.client != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		explainSQL := "EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) " + req.Query
		var planJSONStr string
		err := s.client.Pool().QueryRow(ctx, explainSQL).Scan(&planJSONStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to execute EXPLAIN on database: "+err.Error())
			return
		}

		result, err := explain.ParsePlanJSON([]byte(planJSONStr))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to parse returned execution plan: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}

	// Fallback to demo plan
	writeJSON(w, http.StatusOK, model.GetDemoExplainResult())
}

func (s *Server) handleIndexesAdvisor(w http.ResponseWriter, r *http.Request) {
	if s.demoMode {
		writeJSON(w, http.StatusOK, model.GetDemoRecommendations())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tables, err := s.client.GetTableStats(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve table stats for index advisor: "+err.Error())
		return
	}

	queries, err := s.client.GetSlowQueries(ctx, 100, "total")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve query stats for index advisor: "+err.Error())
		return
	}

	recs := synthesizer.AdviseIndexes(tables, queries)
	writeJSON(w, http.StatusOK, recs)
}

func (s *Server) handleAutovacuum(w http.ResponseWriter, r *http.Request) {
	if s.demoMode {
		workers, wrap := model.GetDemoAutovacuum()
		res := map[string]any{
			"workers":    workers,
			"wraparound": wrap,
		}
		writeJSON(w, http.StatusOK, res)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	workers, err := watchdog.GetAutovacuumWorkers(ctx, s.client)
	if err != nil {
		workers = []model.AutovacuumWorker{}
	}

	wrap, err := watchdog.CheckWraparound(ctx, s.client)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to inspect wraparound safety: "+err.Error())
		return
	}

	res := map[string]any{
		"workers":    workers,
		"wraparound": wrap,
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleBloat(w http.ResponseWriter, r *http.Request) {
	if s.demoMode {
		writeJSON(w, http.StatusOK, model.GetDemoBloat())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	bloat, err := watchdog.EstimateTableBloat(ctx, s.client)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to estimate table bloat: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, bloat)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}
