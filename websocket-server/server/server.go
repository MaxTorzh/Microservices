package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"websocket_server/client"
	"websocket_server/models"

	"github.com/gorilla/websocket"
)

type Server struct {
	clients    map[string]*client.Client
	register   chan *client.Client
	unregister chan *client.Client
	broadcast  chan *models.Message
	mu         sync.RWMutex
	upgrader   websocket.Upgrader
}

func NewServer() *Server {
	return &Server{
		clients:    make(map[string]*client.Client),
		register:   make(chan *client.Client),
		unregister: make(chan *client.Client),
		broadcast:  make(chan *models.Message, 256),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client.ID] = client
			s.mu.Unlock()
			
			welcomeMsg := models.NewSystemMessage("Welcome to the server!")
			client.Send <- welcomeMsg
			
			joinMsg := models.NewSystemMessage("Client " + client.ID + " joined the chat")
			s.Broadcast(joinMsg)
			
			log.Printf("Клиент подключен: %s (Всего клиентов: %d)", client.ID, len(s.clients))
			
		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client.ID]; ok {
				delete(s.clients, client.ID)
				close(client.Send)
			}
			s.mu.Unlock()
			
			leaveMsg := models.NewSystemMessage("Client " + client.ID + " left the chat")
			s.Broadcast(leaveMsg)
			
			log.Printf("Клиент отключен: %s (Всего клиентов: %d)", client.ID, len(s.clients))
			
		case message := <-s.broadcast:
			s.mu.RLock()
			for _, client := range s.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(s.clients, client.ID)
				}
			}
			s.mu.RUnlock()
		}
	}
}

func (s *Server) Register(client *client.Client) {
	s.register <- client
}

func (s *Server) Unregister(client *client.Client) {
	s.unregister <- client
}

func (s *Server) Broadcast(message *models.Message) {
	s.broadcast <- message
}

func (s *Server) SendToClient(clientID string, message *models.Message) error {
	s.mu.RLock()
	client, ok := s.clients[clientID]
	s.mu.RUnlock()
	
	if !ok {
		return fmt.Errorf("client not found: %s", clientID)
	}
	
	select {
	case client.Send <- message:
		return nil
	default:
		return fmt.Errorf("client send channel is full")
	}
}

func (s *Server) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

func (s *Server) GetClients() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	clients := make([]string, 0, len(s.clients))
	for id := range s.clients {
		clients = append(clients, id)
	}
	return clients
}