package messaging

import (
	"context"
	"fmt"

	kafka "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	topic  string
}

func NewProducer(brokers []string, topic string) *Producer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: w,
		topic:  topic,
	}
}

func (p *Producer) Send(ctx context.Context, key []byte, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
	}

	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		zap.L().Error("Kafka message could not be sent", zap.Error(err))
		return fmt.Errorf("Kafka message: %w", err)
	}
	zap.L().Info("Kafka sent the message", zap.String("topic", p.topic))
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
