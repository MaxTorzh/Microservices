package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"user-service/internal/domain"
)

type Producer struct {
    producer sarama.SyncProducer
    topic    string
    logger   *zap.Logger
}

func NewProducer(brokers []string, topic string, logger *zap.Logger) (*Producer, error) {
    config := sarama.NewConfig()
    config.Producer.RequiredAcks = sarama.WaitForAll
    config.Producer.Retry.Max = 5
    config.Producer.Return.Successes = true
    config.Producer.Partitioner = sarama.NewRandomPartitioner
    config.Version = sarama.V2_6_0_0
    
    producer, err := sarama.NewSyncProducer(brokers, config)
    if err != nil {
        return nil, fmt.Errorf("failed to create producer: %w", err)
    }
    
    logger.Info("Kafka producer created", 
        zap.Strings("brokers", brokers),
        zap.String("topic", topic),
    )
    
    return &Producer{
        producer: producer,
        topic:    topic,
        logger:   logger,
    }, nil
}

func (p *Producer) SendUserEvent(event domain.UserEvent) error {
    eventBytes, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }
    
    msg := &sarama.ProducerMessage{
        Topic: p.topic,
        Key:   sarama.StringEncoder(event.UserID),
        Value: sarama.ByteEncoder(eventBytes),
        Headers: []sarama.RecordHeader{
            {
                Key:   []byte("event_type"),
                Value: []byte(event.EventType),
            },
            {
                Key:   []byte("timestamp"),
                Value: []byte(event.Timestamp.String()),
            },
        },
    }
    
    partition, offset, err := p.producer.SendMessage(msg)
    if err != nil {
        return fmt.Errorf("failed to send message: %w", err)
    }
    
    p.logger.Debug("Event sent to Kafka",
        zap.String("event_type", event.EventType),
        zap.String("user_id", event.UserID),
        zap.Int32("partition", partition),
        zap.Int64("offset", offset),
    )
    
    return nil
}

func (p *Producer) Close() error {
    return p.producer.Close()
}