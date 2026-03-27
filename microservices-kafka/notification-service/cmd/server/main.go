package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"notification-service/internal/handler/notification"
	"notification-service/internal/kafka"
	"notification-service/internal/service"
	"notification-service/pkg/logger"
)

func main() {
    // Загрузка в .env файл
    if err := godotenv.Load(); err != nil {
        log.Println("Warning: .env file not found")
    }
    
    // Инициализация логгера
    logLevel := os.Getenv("LOG_LEVEL")
    if logLevel == "" {
        logLevel = "info"
    }
    
    zapLogger, err := logger.NewLogger(logLevel)
    if err != nil {
        log.Fatalf("Failed to create logger: %v", err)
    }
    defer zapLogger.Sync()
    
    zapLogger.Info("Starting Notification Service")
    
    // Настройка HTTP порта
    port := os.Getenv("HTTP_PORT")
    if port == "" {
        port = "8081"
    }
    
    // Настройка Kafka
    kafkaBroker := os.Getenv("KAFKA_BROKER")
    if kafkaBroker == "" {
        kafkaBroker = "localhost:9092"
    }
    kafkaTopic := os.Getenv("KAFKA_TOPIC")
    if kafkaTopic == "" {
        kafkaTopic = "user-events"
    }
    kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")
    if kafkaGroupID == "" {
        kafkaGroupID = "notification-group"
    }
    
    // Создание notifier
    notifier := service.NewNotifier(zapLogger)
    
    // Создание HTTP handler
    httpHandler := notification.NewHandler(zapLogger, notifier)
    
    // Создание Kafka consumer
    consumer, err := kafka.NewConsumer([]string{kafkaBroker}, kafkaGroupID, kafkaTopic, notifier, httpHandler, zapLogger)
    if err != nil {
        zapLogger.Fatal("Failed to create Kafka consumer", zap.Error(err))
    }
    defer consumer.Close()
    
    zapLogger.Info("Connected to Kafka", 
        zap.String("broker", kafkaBroker),
        zap.String("topic", kafkaTopic),
        zap.String("group_id", kafkaGroupID),
    )
    
    // Запуск consumer в горутине
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        if err := consumer.Start(ctx); err != nil && err != context.Canceled {
            zapLogger.Error("Consumer error", zap.Error(err))
        }
    }()
    
    // Настройка HTTP маршрутов
    mux := http.NewServeMux()
    httpHandler.RegisterRoutes(mux)
    
    // Дополнительный эндпоинт для отладки
    mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(fmt.Sprintf(`{
            "service": "notification-service",
            "version": "1.0.0",
            "endpoints": [
                "/health",
                "/ready", 
                "/stats",
                "/metrics",
                "/last-notification"
            ],
            "kafka": {
                "topic": "%s",
                "group": "%s"
            }
        }`, kafkaTopic, kafkaGroupID)))
    })
    
    server := &http.Server{
        Addr:         ":" + port,
        Handler:      mux,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // Graceful shutdown
    sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()
    
    go func() {
        zapLogger.Info("Notification Service starting", zap.String("port", port))
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            zapLogger.Fatal("Failed to start server", zap.Error(err))
        }
    }()
    
    <-sigCtx.Done()
    zapLogger.Info("Shutting down Notification Service...")
    
    // Контекст для graceful shutdown
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()
    
    // Остановка Kafka consumer
    cancel()
    
    // Остановка HTTP сервер
    if err := server.Shutdown(shutdownCtx); err != nil {
        zapLogger.Error("Server shutdown error", zap.Error(err))
    }
    
    zapLogger.Info("Notification Service stopped")
}