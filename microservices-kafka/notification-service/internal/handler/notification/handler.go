package notification

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"notification-service/internal/service"
)

type Handler struct {
    logger           *zap.Logger
    notifier         *service.Notifier
    processedCount   atomic.Int64
    errorCount       atomic.Int64
    lastEventType    atomic.Value
    lastEventTime    atomic.Value
    startTime        time.Time  
}

func NewHandler(logger *zap.Logger, notifier *service.Notifier) *Handler {
    h := &Handler{
        logger:    logger,
        notifier:  notifier,
        startTime: time.Now(),
    }
    h.lastEventType.Store("none")
    h.lastEventTime.Store(time.Time{})
    return h
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("GET /health", h.Health)
    mux.HandleFunc("GET /ready", h.Ready)
    mux.HandleFunc("GET /stats", h.Stats)
    mux.HandleFunc("GET /metrics", h.Metrics)
    mux.HandleFunc("GET /last-notification", h.LastNotification)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "healthy",
        "service": "notification-service",
    })
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ready"))
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
    sent, errors, lastNotification := h.notifier.GetStats()
    
    lastEvent := h.lastEventType.Load()
    lastEventTime := h.lastEventTime.Load().(time.Time)
    
    stats := map[string]interface{}{
        "status":                  "running",
        "service":                 "notification-service",
        "processed_count":         h.processedCount.Load(),
        "error_count":             h.errorCount.Load(),
        "notifier_sent_count":     sent,
        "notifier_error_count":    errors,
        "last_event_type":         lastEvent,
        "last_event_time":         lastEventTime.Format(time.RFC3339),
        "last_notification":       lastNotification,
        "uptime_seconds":          time.Since(h.startTime).Seconds(),
        "start_time":              h.startTime.Format(time.RFC3339),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
    sent, errors, _ := h.notifier.GetStats()
    
    metrics := map[string]interface{}{
        "notification_processed_total": h.processedCount.Load(),
        "notification_errors_total":    h.errorCount.Load(),
        "notifier_sent_total":          sent,
        "notifier_errors_total":        errors,
        "uptime_seconds":               time.Since(h.startTime).Seconds(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(metrics)
}

func (h *Handler) LastNotification(w http.ResponseWriter, r *http.Request) {
    lastNotification := h.notifier.GetLastNotification()
    
    w.Header().Set("Content-Type", "application/json")
    if lastNotification.IsEmpty() {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]string{
            "message": "No notifications sent yet",
        })
        return
    }
    
    json.NewEncoder(w).Encode(lastNotification)
}

func (h *Handler) IncrementProcessed() {
    h.processedCount.Add(1)
}

func (h *Handler) IncrementError() {
    h.errorCount.Add(1)
}

func (h *Handler) SetLastEvent(eventType string) {
    h.lastEventType.Store(eventType)
    h.lastEventTime.Store(time.Now())
}

// GetUptime - время работы сервиса
func (h *Handler) GetUptime() time.Duration {
    return time.Since(h.startTime)
}

// GetStartTime - время запуска сервиса
func (h *Handler) GetStartTime() time.Time {
    return h.startTime
}