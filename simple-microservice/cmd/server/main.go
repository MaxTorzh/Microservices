package main

import (
	"log"
	"net/http"
	"os"

	"simple_microservice/internal/handler/user"
	"simple_microservice/internal/repository/memory"
	"simple_microservice/internal/service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	repo := memory.NewRepository()
	userService := service.NewUserService(repo)
	handler := user.NewHandler(userService)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  POST   /users          - Create user")
	log.Printf("  GET    /users/{id}     - Get user by ID")
	log.Printf("  GET    /users          - Get all users")
	log.Printf("  PUT    /users/{id}     - Update user")
	log.Printf("  DELETE /users/{id}     - Delete user")
	log.Printf("  GET    /health         - Health check")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
