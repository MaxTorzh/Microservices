package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"monitoring-app/internal/config"
	"monitoring-app/internal/handler"
	"monitoring-app/internal/logger"
	"monitoring-app/internal/middleware"
)

func main() {
    cfg := config.Load()
    
    logger.Init(cfg.GetLogLevel())
    logger.Infof("Starting Monitoring Application on port %s", cfg.GetPort())
    
    h := handler.NewHandler()
    
    http.HandleFunc("/", middleware.Metrics(h.HomeHandler))
    http.HandleFunc("/slow", middleware.Metrics(h.SlowHandler))
    http.HandleFunc("/error", middleware.Metrics(h.ErrorHandler))
    http.HandleFunc("/cpu", middleware.Metrics(h.CPUHandler))
    http.HandleFunc("/memory", middleware.Metrics(h.MemoryHandler))
    http.HandleFunc("/health", h.HealthHandler)
    http.Handle("/metrics", promhttp.Handler())
    
    if err := http.ListenAndServe(cfg.GetPort(), nil); err != nil {
        logger.Fatalf("Server failed: %v", err)
    }
}