package server

import (
	"context"
	"encoding/json"
	"fmt"
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
	s.mux.HandleFunc("POST /api/explain/diff", s.handleExplainDiff)
	s.mux.HandleFunc("GET /api/indexes/advisor", s.handleIndexesAdvisor)
	s.mux.HandleFunc("GET /api/indexes/unused", s.handleIndexesUnused)
	s.mux.HandleFunc("POST /api/indexes/simulate", s.handleIndexesSimulate)
	s.mux.HandleFunc("GET /api/activity/locks", s.handleActivityLocks)
	s.mux.HandleFunc("GET /api/autovacuum", s.handleAutovacuum)
	s.mux.HandleFunc("GET /api/bloat", s.handleBloat)
	s.mux.HandleFunc("GET /api/report/markdown", s.handleReportMarkdown)
	s.mux.HandleFunc("GET /api/config/tuning", s.handleConfigTuning)
	s.mux.HandleFunc("POST /api/config/tuning", s.handleConfigTuning)
	s.mux.HandleFunc("GET /api/schema/health", s.handleSchemaHealth)
	s.mux.HandleFunc("GET /api/connections/pool", s.handleConnectionPool)
	s.mux.HandleFunc("POST /api/queries/anti-patterns", s.handleQueryAntiPatterns)
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

	unused, err := s.client.GetUnusedIndexes(ctx)
	if err == nil {
		overview.UnusedIndexesCount = len(unused)
	}

	locks, err := s.client.GetLockContention(ctx)
	if err == nil {
		overview.ActiveLocksCount = locks.ActiveLocksCount
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

	if strings.TrimSpace(req.Plan) != "" {
		result, err := explain.ParsePlanJSON([]byte(req.Plan))
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to parse EXPLAIN plan: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}

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

	writeJSON(w, http.StatusOK, model.GetDemoExplainResult())
}

type explainDiffRequest struct {
	Baseline string `json:"baseline"`
	Current  string `json:"current"`
}

func (s *Server) handleExplainDiff(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var req explainDiffRequest
	_ = json.Unmarshal(body, &req)

	if strings.TrimSpace(req.Baseline) == "" || strings.TrimSpace(req.Current) == "" {
		// Return demo plan diff if inputs are empty
		writeJSON(w, http.StatusOK, model.GetDemoPlanDiff())
		return
	}

	baseRes, err := explain.ParsePlanJSON([]byte(req.Baseline))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse baseline plan: "+err.Error())
		return
	}

	currRes, err := explain.ParsePlanJSON([]byte(req.Current))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse current plan: "+err.Error())
		return
	}

	diff := explain.ComparePlans(baseRes, currRes)
	writeJSON(w, http.StatusOK, diff)
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
		writeError(w, http.StatusInternalServerError, "failed to retrieve table stats: "+err.Error())
		return
	}

	queries, err := s.client.GetSlowQueries(ctx, 100, "total")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve query stats: "+err.Error())
		return
	}

	recs := synthesizer.AdviseIndexes(tables, queries)
	writeJSON(w, http.StatusOK, recs)
}

func (s *Server) handleIndexesUnused(w http.ResponseWriter, r *http.Request) {
	if s.demoMode {
		writeJSON(w, http.StatusOK, model.GetDemoUnusedIndexes())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	unused, err := s.client.GetUnusedIndexes(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query unused indexes: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, unused)
}

type simulateRequest struct {
	DDL   string `json:"ddl"`
	Query string `json:"query"`
}

func (s *Server) handleIndexesSimulate(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req simulateRequest
	_ = json.Unmarshal(body, &req)

	if s.demoMode || strings.TrimSpace(req.Query) == "" || strings.TrimSpace(req.DDL) == "" {
		writeJSON(w, http.StatusOK, model.GetDemoHypoSimulation())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := synthesizer.SimulateHypoPG(ctx, s.client, req.DDL, req.Query)
	if err != nil {
		writeError(w, http.StatusBadRequest, "simulation failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleActivityLocks(w http.ResponseWriter, r *http.Request) {
	if s.demoMode {
		writeJSON(w, http.StatusOK, model.GetDemoLockContention())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	locks, err := s.client.GetLockContention(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query lock contention: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, locks)
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

func (s *Server) handleReportMarkdown(w http.ResponseWriter, r *http.Request) {
	var queries []model.QueryStat
	var recs []model.IndexRecommendation
	var unused []model.UnusedIndex
	var bloat []model.BloatInfo

	if s.demoMode {
		queries = model.GetDemoQueries()
		recs = model.GetDemoRecommendations()
		unused = model.GetDemoUnusedIndexes()
		bloat = model.GetDemoBloat()
	} else {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		queries, _ = s.client.GetSlowQueries(ctx, 10, "total")
		tables, _ := s.client.GetTableStats(ctx)
		recs = synthesizer.AdviseIndexes(tables, queries)
		unused, _ = s.client.GetUnusedIndexes(ctx)
		bloat, _ = watchdog.EstimateTableBloat(ctx, s.client)
	}

	var md strings.Builder
	md.WriteString("# IndexLens PostgreSQL Performance & Diagnostic Report\n\n")
	md.WriteString(fmt.Sprintf("*Generated on: %s*\n\n", time.Now().UTC().Format(time.RFC1123)))

	md.WriteString("## 1. Top Identified Missing Indexes (Action Items)\n\n")
	for i, r := range recs {
		md.WriteString(fmt.Sprintf("### [%d] Table: `%s` (Priority: %s)\n", i+1, r.TableName, r.Priority))
		md.WriteString(fmt.Sprintf("- **Reason**: %s\n", r.Reason))
		md.WriteString(fmt.Sprintf("- **Estimated Size**: %s | **Write Penalty**: %s\n", r.EstimatedSizePretty, r.WritePenalty))
		md.WriteString(fmt.Sprintf("```sql\n%s\n```\n\n", r.DDL))
	}

	if len(unused) > 0 {
		md.WriteString("## 2. Unused & Redundant Indexes (Cleanup Candidates)\n\n")
		for i, u := range unused {
			md.WriteString(fmt.Sprintf("### [%d] Index: `%s` on `%s` (%s)\n", i+1, u.IndexName, u.TableName, u.IndexSizePretty))
			md.WriteString(fmt.Sprintf("- **Reason**: %s\n", u.Reason))
			md.WriteString(fmt.Sprintf("```sql\n%s\n```\n\n", u.DropDDL))
		}
	}

	md.WriteString("## 3. High Latency Queries\n\n")
	md.WriteString("| Calls | Mean Latency | Total Duration | Cache Hit % | Query Snippet |\n")
	md.WriteString("|---|---|---|---|---|\n")
	for _, q := range queries {
		snippet := q.Query
		if len(snippet) > 60 {
			snippet = snippet[:57] + "..."
		}
		md.WriteString(fmt.Sprintf("| %d | %.1f ms | %.1f s | %.1f%% | `%s` |\n",
			q.Calls, q.MeanExecTime, q.TotalExecTime/1000.0, q.CacheHitRatio, strings.ReplaceAll(snippet, "|", "/")))
	}
	md.WriteString("\n")

	md.WriteString("## 4. Table Bloat & Dead Tuples\n\n")
	md.WriteString("| Relation | Total Size | Dead Tuples | Dead % | Estimated Bloat |\n")
	md.WriteString("|---|---|---|---|---|\n")
	for _, b := range bloat {
		md.WriteString(fmt.Sprintf("| `%s` | %s | %d | %.1f%% | %s |\n",
			b.TableName, b.SizePretty, b.DeadTuples, b.DeadTupleRatio, b.BloatPretty))
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="indexlens_report.md"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(md.String()))
}

type tuningRequest struct {
	TotalRAMGB   float64 `json:"totalRamGb"`
	VCPUs        int     `json:"vcpus"`
	DiskType     string  `json:"diskType"`
	WorkloadType string  `json:"workloadType"`
}

func (s *Server) handleConfigTuning(w http.ResponseWriter, r *http.Request) {
	var profile model.TuningProfile
	profile.TotalRAMGB = 32.0
	profile.VCPUs = 8
	profile.DiskType = "NVMe"
	profile.WorkloadType = "OLTP"

	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err == nil && len(body) > 0 {
			var req tuningRequest
			if err := json.Unmarshal(body, &req); err == nil {
				if req.TotalRAMGB > 0 {
					profile.TotalRAMGB = req.TotalRAMGB
				}
				if req.VCPUs > 0 {
					profile.VCPUs = req.VCPUs
				}
				if req.DiskType != "" {
					profile.DiskType = req.DiskType
				}
				if req.WorkloadType != "" {
					profile.WorkloadType = req.WorkloadType
				}
			}
		}
	} else {
		// Read from query params
		if ramStr := r.URL.Query().Get("ramGb"); ramStr != "" {
			if v, err := strconv.ParseFloat(ramStr, 64); err == nil && v > 0 {
				profile.TotalRAMGB = v
			}
		}
		if cpuStr := r.URL.Query().Get("vcpus"); cpuStr != "" {
			if v, err := strconv.Atoi(cpuStr); err == nil && v > 0 {
				profile.VCPUs = v
			}
		}
		if disk := r.URL.Query().Get("disk"); disk != "" {
			profile.DiskType = disk
		}
		if wl := r.URL.Query().Get("workload"); wl != "" {
			profile.WorkloadType = wl
		}
	}

	if s.demoMode || s.client == nil {
		report := model.GetDemoTuningReport(profile)
		writeJSON(w, http.StatusOK, report)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	liveSettings, err := s.client.GetLiveSettings(ctx)
	if err != nil {
		// Fallback to calculation without live settings
		liveSettings = make(map[string]string)
	}

	report := poller.CalculateTuning(profile, liveSettings)
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleSchemaHealth(w http.ResponseWriter, r *http.Request) {
	if s.demoMode || s.client == nil {
		writeJSON(w, http.StatusOK, model.GetDemoSchemaHealth())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	report, err := s.client.GetSchemaHealthReport(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to audit schema health: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleConnectionPool(w http.ResponseWriter, r *http.Request) {
	vcpus := 4
	if cpuStr := r.URL.Query().Get("vcpus"); cpuStr != "" {
		if v, err := strconv.Atoi(cpuStr); err == nil && v > 0 {
			vcpus = v
		}
	}

	if s.demoMode || s.client == nil {
		writeJSON(w, http.StatusOK, model.GetDemoConnectionPoolHealth())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	poolHealth, err := s.client.GetConnectionPoolHealth(ctx, vcpus)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to inspect connection pool: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, poolHealth)
}

type antiPatternRequest struct {
	Query string `json:"query"`
}

func (s *Server) handleQueryAntiPatterns(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var req antiPatternRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	patterns := explain.AnalyzeSQLAntiPatterns(req.Query)
	writeJSON(w, http.StatusOK, patterns)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}
