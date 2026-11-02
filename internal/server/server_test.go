package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServerEndpointsInDemoMode(t *testing.T) {
	srv := NewServer(nil, true)
	handler := srv.Handler()

	tests := []struct {
		name       string
		method     string
		url        string
		body       string
		wantStatus int
		wantSubstr string
	}{
		{
			name:       "Health check",
			method:     "GET",
			url:        "/api/health",
			wantStatus: http.StatusOK,
			wantSubstr: `"status":"healthy"`,
		},
		{
			name:       "Overview",
			method:     "GET",
			url:        "/api/overview",
			wantStatus: http.StatusOK,
			wantSubstr: `"demoMode":true`,
		},
		{
			name:       "Queries",
			method:     "GET",
			url:        "/api/queries",
			wantStatus: http.StatusOK,
			wantSubstr: `orders`,
		},
		{
			name:       "Explain demo fallback",
			method:     "POST",
			url:        "/api/explain",
			body:       `{}`,
			wantStatus: http.StatusOK,
			wantSubstr: `HEAVY_SEQ_SCAN`,
		},
		{
			name:       "Explain diff fallback",
			method:     "POST",
			url:        "/api/explain/diff",
			body:       `{}`,
			wantStatus: http.StatusOK,
			wantSubstr: `IMPROVED`,
		},
		{
			name:       "Indexes advisor",
			method:     "GET",
			url:        "/api/indexes/advisor",
			wantStatus: http.StatusOK,
			wantSubstr: `CREATE INDEX CONCURRENTLY`,
		},
		{
			name:       "Unused indexes",
			method:     "GET",
			url:        "/api/indexes/unused",
			wantStatus: http.StatusOK,
			wantSubstr: `DROP INDEX CONCURRENTLY`,
		},
		{
			name:       "Hypo simulation",
			method:     "POST",
			url:        "/api/indexes/simulate",
			body:       `{}`,
			wantStatus: http.StatusOK,
			wantSubstr: `Planner verified`,
		},
		{
			name:       "Activity and locks",
			method:     "GET",
			url:        "/api/activity/locks",
			wantStatus: http.StatusOK,
			wantSubstr: `rootBlockers`,
		},
		{
			name:       "Autovacuum",
			method:     "GET",
			url:        "/api/autovacuum",
			wantStatus: http.StatusOK,
			wantSubstr: `wraparound`,
		},
		{
			name:       "Bloat",
			method:     "GET",
			url:        "/api/bloat",
			wantStatus: http.StatusOK,
			wantSubstr: `audit_logs`,
		},
		{
			name:       "Markdown report",
			method:     "GET",
			url:        "/api/report/markdown",
			wantStatus: http.StatusOK,
			wantSubstr: `# IndexLens PostgreSQL Performance`,
		},
		{
			name:       "Config tuning",
			method:     "GET",
			url:        "/api/config/tuning?ramGb=32&vcpus=8&disk=NVMe&workload=OLTP",
			wantStatus: http.StatusOK,
			wantSubstr: `shared_buffers`,
		},
		{
			name:       "Schema health",
			method:     "GET",
			url:        "/api/schema/health",
			wantStatus: http.StatusOK,
			wantSubstr: `unindexedFks`,
		},
		{
			name:       "Connection pool",
			method:     "GET",
			url:        "/api/connections/pool",
			wantStatus: http.StatusOK,
			wantSubstr: `optimalPoolSize`,
		},
		{
			name:       "Query anti-patterns POST",
			method:     "POST",
			url:        "/api/queries/anti-patterns",
			body:       `{"query":"SELECT * FROM users WHERE lower(email) = 'test@example.com'"}`,
			wantStatus: http.StatusOK,
			wantSubstr: `FUNCTION_ON_INDEX_COLUMN`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *strings.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			} else {
				bodyReader = strings.NewReader("")
			}

			req := httptest.NewRequest(tt.method, tt.url, bodyReader)
			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			respBody := rec.Body.String()
			if !strings.Contains(respBody, tt.wantSubstr) {
				t.Errorf("expected response to contain %q, got %s", tt.wantSubstr, respBody)
			}
		})
	}
}
