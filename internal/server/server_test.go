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
			name:       "Indexes advisor",
			method:     "GET",
			url:        "/api/indexes/advisor",
			wantStatus: http.StatusOK,
			wantSubstr: `CREATE INDEX CONCURRENTLY`,
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
