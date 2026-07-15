package middleware

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type MetricsCollector struct {
	totalRequests  int64
	totalErrors    int64
	totalLatencyMs int64
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

func (m *MetricsCollector) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			atomic.AddInt64(&m.totalRequests, 1)

			rec := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			if rec.statusCode >= 500 {
				atomic.AddInt64(&m.totalErrors, 1)
			}
			elapsed := time.Since(start).Milliseconds()
			atomic.AddInt64(&m.totalLatencyMs, elapsed)
		})
	}
}

func (m *MetricsCollector) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqs := atomic.LoadInt64(&m.totalRequests)
		errs := atomic.LoadInt64(&m.totalErrors)
		lat := atomic.LoadInt64(&m.totalLatencyMs)
		avgLat := float64(0)
		if reqs > 0 {
			avgLat = float64(lat) / float64(reqs)
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP http_requests_total Total number of HTTP requests processed.\n")
		fmt.Fprintf(w, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(w, "http_requests_total %d\n\n", reqs)

		fmt.Fprintf(w, "# HELP http_server_errors_total Total number of 5xx HTTP errors.\n")
		fmt.Fprintf(w, "# TYPE http_server_errors_total counter\n")
		fmt.Fprintf(w, "http_server_errors_total %d\n\n", errs)

		fmt.Fprintf(w, "# HELP http_avg_latency_ms Average HTTP request latency in milliseconds.\n")
		fmt.Fprintf(w, "# TYPE http_avg_latency_ms gauge\n")
		fmt.Fprintf(w, "http_avg_latency_ms %.2f\n", avgLat)
	}
}
