package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Счетчик запросов
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_requests_total",
			Help: "total_number_requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Гистограмма времени ответа
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"method", "endpoint"},
	)

	// Датчик активных соединений
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_active_connections",
			Help: "Number of active connections",
		},
	)

	// Счетчик ошибок
	ErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "app_errors_total",
			Help: "Total number of errors",
		},
	)
)

func IncRequestsTotal(method, endpoint, status string) {
    RequestsTotal.WithLabelValues(method, endpoint, status).Inc()
}

func ObserveRequestDuration(method, endpoint string, duration float64) {
    RequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

func IncActiveConnections() {
    ActiveConnections.Inc()
}

func DecActiveConnections() {
    ActiveConnections.Dec()
}

func IncErrorsTotal() {
    ErrorsTotal.Inc()
}