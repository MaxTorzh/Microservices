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
    
    "user-service/internal/handler/user"
    "user-service/internal/kafka"
    "user-service/internal/repository/postgres"
    "user-service/internal/service"
    "user-service/pkg/logger"
)

func main() {
    // Загрузка .env файл
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
    
    zapLogger.Info("Starting User Service")
    
    // Настройка HTTP порта
    port := os.Getenv("HTTP_PORT")
    if port == "" {
        port = "8080"
    }
    
    // Настройка PostgreSQL
    dbHost := os.Getenv("DB_HOST")
    if dbHost == "" {
        dbHost = "localhost"
    }
    dbPort := os.Getenv("DB_PORT")
    if dbPort == "" {
        dbPort = "5432"
    }
    dbUser := os.Getenv("DB_USER")
    if dbUser == "" {
        dbUser = "postgres"
    }
    dbPassword := os.Getenv("DB_PASSWORD")
    if dbPassword == "" {
        dbPassword = "postgres"
    }
    dbName := os.Getenv("DB_NAME")
    if dbName == "" {
        dbName = "user_service"
    }
    
    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        dbHost, dbPort, dbUser, dbPassword, dbName)
    
    // Создание репозитория
    repo, err := postgres.NewRepository(connStr)
    if err != nil {
        zapLogger.Fatal("Failed to connect to database", zap.Error(err))
    }
    defer repo.Close()
    
    zapLogger.Info("Connected to PostgreSQL", zap.String("database", dbName))
    
    // Настройка Kafka
    kafkaBroker := os.Getenv("KAFKA_BROKER")
    if kafkaBroker == "" {
        kafkaBroker = "localhost:9092"
    }
    kafkaTopic := os.Getenv("KAFKA_TOPIC")
    if kafkaTopic == "" {
        kafkaTopic = "user-events"
    }
    
    // Создание producer
    producer, err := kafka.NewProducer([]string{kafkaBroker}, kafkaTopic, zapLogger)
    if err != nil {
        zapLogger.Fatal("Failed to create Kafka producer", zap.Error(err))
    }
    defer producer.Close()
    
    zapLogger.Info("Connected to Kafka", zap.String("broker", kafkaBroker))
    
    // Создание сервисов
    userService := service.NewUserService(repo, producer, zapLogger)
    userHandler := user.NewHandler(userService, zapLogger)
    
    // Настройка HTTP маршрутов
    mux := http.NewServeMux()
    userHandler.RegisterRoutes(mux)
    
    // Health check endpoint
    mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        if err := repo.Ping(); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            w.Write([]byte(fmt.Sprintf(`{"status":"unhealthy","error":"%s"}`, err.Error())))
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"healthy","service":"user-service"}`))
    })
    
    // Ready endpoint
    mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
        if err := repo.Ping(); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ready"))
    })
    
    server := &http.Server{
        Addr:         ":" + port,
        Handler:      mux,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // Graceful shutdown
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()
    
    go func() {
        zapLogger.Info("User Service starting", zap.String("port", port))
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            zapLogger.Fatal("Failed to start server", zap.Error(err))
        }
    }()
    
    <-ctx.Done()
    zapLogger.Info("Shutting down User Service...")
    
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(shutdownCtx); err != nil {
        zapLogger.Error("Server shutdown error", zap.Error(err))
    }
    
    zapLogger.Info("User Service stopped")
}