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

	pb "gRPC/api/proto/user"
	grpcHandler "gRPC/internal/handler/grpc"
	httpHandler "gRPC/internal/handler/http_handler"
	"gRPC/internal/repository/memory"
	"gRPC/internal/service"

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

	repo := memory.NewRepository()
	userService := service.NewUserService(repo)

	httpHandler := httpHandler.NewHandler(userService)
	grpcHandler := grpcHandler.NewUserGrpcHandler(userService)

	// HTTP Server
	httpMux := http.NewServeMux()
	httpHandler.RegisterRoutes(httpMux)

	httpMux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","protocol":"http"}`))
	})

	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: httpMux,
	}

	// gRPC Server
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, grpcHandler)
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