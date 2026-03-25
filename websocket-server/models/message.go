package models

import (
	"encoding/json"
	"time"
)

const (
	MessageTypeText   = "text"
	MessageTypeBinary = "binary"
	MessageTypeSystem = "system"
)

// Message структура для обмена сообщениями
type Message struct {
	Type      string      `json:"type"`
	ID        string      `json:"id"`
	ClientID  string      `json:"client_id"`
	Content   interface{} `json:"content"`
	Timestamp time.Time   `json:"timestamp"`
}

type TextMessage struct {
	Text string `json:"text"`
}

type BinaryMessage struct {
	Data     string `json:"data"`
	Encoding string `json:"encoding"`
}

func NewTextMessage(clientID, text string) *Message {
	return &Message{
		Type:      MessageTypeText,
		ID:        generateID(),
		ClientID:  clientID,
		Content:   TextMessage{Text: text},
		Timestamp: time.Now(),
	}
}

func NewBinaryMessage(clientID, data, encoding string) *Message {
	return &Message{
		Type:      MessageTypeBinary,
		ID:        generateID(),
		ClientID:  clientID,
		Content:   BinaryMessage{Data: data, Encoding: encoding},
		Timestamp: time.Now(),
	}
}

func NewSystemMessage(content string) *Message {
	return &Message{
		Type:      MessageTypeSystem,
		ID:        generateID(),
		ClientID:  "system",
		Content:   TextMessage{Text: content},
		Timestamp: time.Now(),
	}
}

func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

func generateID() string {
	return time.Now().Format("20060102150405.000000")
}