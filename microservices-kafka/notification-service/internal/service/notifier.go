package service

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"notification-service/internal/domain"
)

type Notifier struct {
    logger         *zap.Logger
    totalSent      atomic.Int64
    totalErrors    atomic.Int64
    lastNotification atomic.Value
}

func NewNotifier(logger *zap.Logger) *Notifier {
    n := &Notifier{
        logger: logger,
    }
    n.lastNotification.Store(&domain.Notification{})
    return n
}

func (n *Notifier) ProcessEvent(event domain.UserEvent) error {
    notification := domain.Notification{
        ID:     uuid.New().String(),
        UserID: event.UserID,
        SentAt: time.Now(),
    }
    
    var notificationType, message string
    
    switch event.EventType {
    case "created":
        notificationType = "welcome"
        message = fmt.Sprintf(
            "Welcome %s! Thank you for registering with email %s.",
            event.Name, event.Email,
        )
        
    case "updated":
        notificationType = "update"
        message = fmt.Sprintf(
            "Your profile has been updated. Name: %s, Email: %s",
            event.Name, event.Email,
        )
        
    case "deleted":
        notificationType = "goodbye"
        message = fmt.Sprintf(
            "Goodbye %s. Your account has been deleted.",
            event.Name,
        )
        
    default:
        n.totalErrors.Add(1)
        return fmt.Errorf("unknown event type: %s", event.EventType)
    }
    
    notification.Type = notificationType
    notification.Message = message
    
    // Имитация отправки уведомления
    startTime := time.Now()
    
    n.lastNotification.Store(&notification)
    n.totalSent.Add(1)
    
    n.logger.Info("Notification sent",
        zap.String("notification_id", notification.ID),
        zap.String("type", notification.Type),
        zap.String("user_id", notification.UserID),
        zap.String("message", truncate(message, 100)),
        zap.Duration("duration", time.Since(startTime)),
        zap.Time("sent_at", notification.SentAt),
    )
    
    return nil
}

func (n *Notifier) GetStats() (sent int64, errors int64, lastNotification *domain.Notification) {
    sent = n.totalSent.Load()
    errors = n.totalErrors.Load()
    lastNotification = n.lastNotification.Load().(*domain.Notification)
    return
}

func (n *Notifier) GetLastNotification() *domain.Notification {
    return n.lastNotification.Load().(*domain.Notification)
}

func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen] + "..."
}