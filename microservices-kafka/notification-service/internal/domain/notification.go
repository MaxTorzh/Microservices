package domain

import (
	"fmt"
	"time"
)

type UserEvent struct {
    EventType string    `json:"event_type"`
    UserID    string    `json:"user_id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    Timestamp time.Time `json:"timestamp"`
}

type Notification struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Type      string    `json:"type"`
    Message   string    `json:"message"`
    SentAt    time.Time `json:"sent_at"`
}

// Возврат строкового представления уведомления
func (n Notification) String() string {
    return fmt.Sprintf("Notification{id=%s, user_id=%s, type=%s, sent_at=%s}",
        n.ID, n.UserID, n.Type, n.SentAt.Format(time.RFC3339))
}

// Проверка на пустое уведомление
func (n Notification) IsEmpty() bool {
    return n.ID == ""
}