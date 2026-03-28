package handler

import (
	"math/rand"
	"net/http"
	"time"

	"monitoring-app/internal/logger"
	"monitoring-app/internal/metrics"
	"monitoring-app/pkg/response"
)

type Handler struct{}

func NewHandler() *Handler {
    return &Handler{}
}

// Простой эндпоинт
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
    response.Success(w, map[string]string{
        "message": "Hello from monitoring app!",
    })
}

// Эндпоинт с задержкой
func (h *Handler) SlowHandler(w http.ResponseWriter, r *http.Request) {
    delay := time.Duration(10+rand.Intn(490)) * time.Millisecond
    time.Sleep(delay)
    
    response.Success(w, map[string]interface{}{
        "message":  "Slow response",
        "delay_ms": delay.Milliseconds(),
    })
}

// Эндпоинт с ошибкой
func (h *Handler) ErrorHandler(w http.ResponseWriter, r *http.Request) {
    metrics.IncErrorsTotal()
    
    logger.WithFields(map[string]interface{}{
        "endpoint": "/error",
        "error":    "simulated error",
    }).Error("Error occurred")
    
    response.Error(w, http.StatusInternalServerError, "Internal server error")
}

// Эндпоинт для CPU нагрузки
func (h *Handler) CPUHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    result := 0
    for i := 0; i < 10000000; i++ {
        result += i
    }
    duration := time.Since(start)
    
    logger.WithFields(map[string]interface{}{
        "endpoint":    "/cpu",
        "duration_ms": duration.Milliseconds(),
        "result":      result,
    }).Info("CPU task completed")
    
    response.Success(w, map[string]interface{}{
        "message":     "CPU task completed",
        "duration_ms": duration.Milliseconds(),
        "result":      result,
    })
}

// Эндпоинт для нагрузки памяти
func (h *Handler) MemoryHandler(w http.ResponseWriter, r *http.Request) {
    data := make([]byte, 10*1024*1024)
    for i := range data {
        data[i] = byte(i % 256)
    }
    
    logger.WithFields(map[string]interface{}{
        "endpoint":     "/memory",
        "allocated_mb": 10,
    }).Info("Memory allocated")
    
    response.Success(w, map[string]interface{}{
        "message": "Memory allocated",
        "size_mb": 10,
    })
}

// Health check
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
    response.Success(w, map[string]string{
        "status":  "healthy",
        "service": "monitoring-app",
    })
}