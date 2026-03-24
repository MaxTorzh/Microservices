package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "JWT/api/proto/user"
	"JWT/internal/auth"
	grpcHandler "JWT/internal/handler/grpc"
	httpHandler "JWT/internal/handler/http_handler"
	"JWT/internal/repository/memory"
	"JWT/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}

	// Ключ для JWT (учебный проект, поэтому не в .env)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}

	// JWT менеджер (токен живет 24 часа)
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)

	// Интерсептор аутентификации
	authInterceptor := auth.NewAuthInterceptor(jwtManager)

	// Инициализация зависимостей
	repo := memory.NewRepository()
	userService := service.NewUserService(repo)

	// HTTP Handler (REST)
	httpHandlerInst := httpHandler.NewHandler(userService, jwtManager)

	// gRPC Handler
	grpcHandlerInst := grpcHandler.NewUserGrpcHandler(userService, jwtManager)

	// HTTP Server
	httpMux := http.NewServeMux()
	httpHandlerInst.RegisterRoutes(httpMux)

	httpMux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","protocol":"http"}`))
	})

	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: httpMux,
	}

	// gRPC Server с интерсептором аутентификации
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor()),
	)
	pb.RegisterUserServiceServer(grpcServer, grpcHandlerInst)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Printf("HTTP server starting on :%s", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	go func() {
		log.Printf("gRPC server starting on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	log.Println("========================================")
	log.Println("User Service Started")
	log.Printf("HTTP:  http://localhost:%s", httpPort)
	log.Printf("gRPC:  localhost:%s", grpcPort)
	log.Println("Authentication: JWT required for protected endpoints")
	log.Println("========================================")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	httpServer.Shutdown(ctx)
	grpcServer.GracefulStop()

	log.Println("Servers stopped")
}