package server

import (
	"log"
	"net/http"
	"time"
)

// withMiddleware adds CORS headers, request timing, and panic recovery.
func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Set permissive CORS headers for local studio development
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[PANIC RECOVER] %v\n", rec)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)

		// Log API requests
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			log.Printf("[API] %s %s (%v)\n", r.Method, r.URL.Path, time.Since(start))
		}
	})
}
