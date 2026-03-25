package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"websocket_server/handler"
	"websocket_server/server"
)

func main() {
	wsServer := server.NewServer()
	
	go wsServer.Run()
	
	wsHandler := handler.NewWebSocketHandler(wsServer)
	
	http.HandleFunc("/", handler.HomeHandler)
	http.Handle("/ws", wsHandler)
	http.HandleFunc("/status", wsHandler.StatusHandler)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}
	
	// Обработка graceful shutdown
	go func() {
		log.Printf("Сервер запущен на порту %s", port)
		log.Printf("WebSocket endpoint: ws://localhost:%s/ws", port)
		log.Printf("Status endpoint: http://localhost:%s/status", port)
		log.Printf("Test client: http://localhost:%s", port)
		
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Сервер завершает работу...")
	if err := httpServer.Close(); err != nil {
		log.Printf("Ошибка при остановке сервера: %v", err)
	}
	log.Println("Сервер остановлен")
}