package middleware

import (
	"net/http"
	"time"

	"monitoring-app/internal/logger"
	"monitoring-app/internal/metrics"
)

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

func Metrics(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        metrics.IncActiveConnections()
        defer metrics.DecActiveConnections()
        
        rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        logger.WithFields(map[string]interface{}{
            "method":      r.Method,
            "path":        r.URL.Path,
            "remote_addr": r.RemoteAddr,
            "user_agent":  r.UserAgent(),
        }).Info("Request started")
        
        next(rw, r)
        
        duration := time.Since(start).Seconds()
        
        metrics.ObserveRequestDuration(r.Method, r.URL.Path, duration)
        metrics.IncRequestsTotal(r.Method, r.URL.Path, string(rune(rw.statusCode)))
        
        logger.WithFields(map[string]interface{}{
            "method":       r.Method,
            "path":         r.URL.Path,
            "status":       rw.statusCode,
            "duration_ms":  duration * 1000,
        }).Info("Request completed")
    }
}