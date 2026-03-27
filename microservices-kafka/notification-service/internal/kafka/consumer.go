package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"notification-service/internal/domain"
	"notification-service/internal/handler/notification"
	"notification-service/internal/service"
)

type Consumer struct {
    consumerGroup sarama.ConsumerGroup
    topic         string
    notifier      *service.Notifier
    handler       *notification.Handler
    logger        *zap.Logger
}

func NewConsumer(brokers []string, groupID, topic string, notifier *service.Notifier, handler *notification.Handler, logger *zap.Logger) (*Consumer, error) {
    config := sarama.NewConfig()
    
    config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
        sarama.NewBalanceStrategyRoundRobin(),
    }
    
    config.Consumer.Offsets.Initial = sarama.OffsetOldest
    config.Consumer.Return.Errors = true
    config.Version = sarama.V2_6_0_0
    
    consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
    if err != nil {
        return nil, fmt.Errorf("failed to create consumer group: %w", err)
    }
    
    logger.Info("Kafka consumer created",
        zap.Strings("brokers", brokers),
        zap.String("group_id", groupID),
        zap.String("topic", topic),
    )
    
    return &Consumer{
        consumerGroup: consumerGroup,
        topic:         topic,
        notifier:      notifier,
        handler:       handler,
        logger:        logger,
    }, nil
}

func (c *Consumer) Start(ctx context.Context) error {
    handler := &ConsumerHandler{
        notifier: c.notifier,
        handler:  c.handler,
        logger:   c.logger,
    }
    
    for {
        if err := c.consumerGroup.Consume(ctx, []string{c.topic}, handler); err != nil {
            c.logger.Error("Error from consumer", zap.Error(err))
            return err
        }
        if ctx.Err() != nil {
            return ctx.Err()
        }
    }
}

func (c *Consumer) Close() error {
    return c.consumerGroup.Close()
}

type ConsumerHandler struct {
    notifier *service.Notifier
    handler  *notification.Handler
    logger   *zap.Logger
}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
    h.logger.Info("Consumer handler setup")
    return nil
}

func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
    h.logger.Info("Consumer handler cleanup")
    return nil
}

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    for message := range claim.Messages() {
        h.logger.Debug("Received message",
            zap.String("topic", message.Topic),
            zap.Int32("partition", message.Partition),
            zap.Int64("offset", message.Offset),
        )
        
        var event domain.UserEvent
        if err := json.Unmarshal(message.Value, &event); err != nil {
            h.logger.Error("Failed to unmarshal message", zap.Error(err))
            h.handler.IncrementError()
            session.MarkMessage(message, "")
            continue
        }
        
        // Обновление статистики
        h.handler.SetLastEvent(event.EventType)
        
        // Обработка события
        startTime := time.Now()
        if err := h.notifier.ProcessEvent(event); err != nil {
            h.logger.Error("Failed to process event",
                zap.Error(err),
                zap.String("event_type", event.EventType),
                zap.String("user_id", event.UserID),
            )
            h.handler.IncrementError()
        } else {
            h.handler.IncrementProcessed()
            h.logger.Info("Event processed successfully",
                zap.String("event_type", event.EventType),
                zap.String("user_id", event.UserID),
                zap.String("email", event.Email),
                zap.Duration("duration", time.Since(startTime)),
            )
        }
        
        session.MarkMessage(message, "")
    }
    return nil
}