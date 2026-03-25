package client

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"websocket_server/models"

	"github.com/gorilla/websocket"
)

// Подключенный клиент
type Client struct {
	ID        string
	Conn      *websocket.Conn
	Send      chan *models.Message
	Server    ServerInterface
	mu        sync.Mutex
	CreatedAt time.Time
}

// Взаимодействие с сервером
type ServerInterface interface {
	Register(client *Client)
	Unregister(client *Client)
	Broadcast(message *models.Message)
	SendToClient(clientID string, message *models.Message) error
}

func NewClient(id string, conn *websocket.Conn, server ServerInterface) *Client {
	return &Client{
		ID:        id,
		Conn:      conn,
		Send:      make(chan *models.Message, 256),
		Server:    server,
		CreatedAt: time.Now(),
	}
}

// Обработка входящих сообщений от клиента
func (c *Client) ReadPump() {
	defer func() {
		c.Server.Unregister(c)
		c.Conn.Close()
	}()

	// Установка лимитов на чтение
	c.Conn.SetReadLimit(512 * 1024) // 512KB
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message models.Message
		
		// Чтение сообщений (JSON или бинарные данные)
		messageType, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Ошибка чтения от клиента %s: %v", c.ID, err)
			}
			break
		}

		// Обработка в зависимости от типа WebSocket сообщения
		switch messageType {
		case websocket.TextMessage:
			// Парсинг JSON сообщения
			if err := json.Unmarshal(data, &message); err != nil {
				log.Printf("Ошибка парсинга JSON от клиента %s: %v", c.ID, err)
				c.sendError("Invalid JSON format")
				continue
			}
			
			// Установка ID клиента если его нет
			if message.ClientID == "" {
				message.ClientID = c.ID
			}
			message.Timestamp = time.Now()
			
			// Отправка на сервер
			c.handleMessage(&message)
			
		case websocket.BinaryMessage:
			// Бинарное сообщение
			binaryMsg := models.NewBinaryMessage(c.ID, string(data), "binary")
			c.Server.Broadcast(binaryMsg)
			log.Printf("Получено бинарное сообщение от %s, размер: %d байт", c.ID, len(data))
		}
	}
}

// WritePump отправляет сообщения клиенту
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Отправляем сообщение в зависимости от его типа
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Ошибка маршалинга сообщения: %v", err)
				continue
			}
			
			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			// Отправка ping для поддержания соединения
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Обработка полученного сообщения
func (c *Client) handleMessage(message *models.Message) {
	switch message.Type {
	case models.MessageTypeText:
		// Рассылка всем клиентам
		c.Server.Broadcast(message)
		log.Printf("Текстовое сообщение от %s: %v", c.ID, message.Content)
		
	case models.MessageTypeBinary:
		c.Server.Broadcast(message)
		log.Printf("Бинарное сообщение от %s", c.ID)
		
	default:
		log.Printf("Неизвестный тип сообщения от %s: %s", c.ID, message.Type)
		c.sendError("Unknown message type")
	}
}

func (c *Client) sendError(errorMsg string) {
	errMsg := models.NewSystemMessage("Error: " + errorMsg)
	select {
	case c.Send <- errMsg:
	default:
	}
}