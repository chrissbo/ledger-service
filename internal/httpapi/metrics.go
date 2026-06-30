package httpapi

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var httpRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests by method, path, and status code.",
	},
	[]string{"method", "path", "status"},
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
}

// MetricsHandler returns the Prometheus metrics handler.
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// instrumentHandler wraps an http.HandlerFunc to record request counts.
func instrumentHandler(path string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next(rec, r)
		httpRequestsTotal.WithLabelValues(r.Method, path, fmt.Sprintf("%d", rec.status)).Inc()
	}
}

// statusRecorder captures the HTTP status code written by the handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}
