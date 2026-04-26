package messaging

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Customer struct {
	reader *kafka.Reader
}

type CustomerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewCustomer(cfg CustomerConfig) *Customer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		Topic:       cfg.Topic,
		GroupID:     cfg.GroupID,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		MaxWait:     2 * time.Second,
		StartOffset: kafka.FirstOffset,
	})
	return &Customer{reader: reader}
}

func (c *Customer) Start(ctx context.Context) {
	zap.L().Info("Kafka customer started",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("group", c.reader.Config().GroupID),
	)

	defer func() {
		if err := c.reader.Close(); err != nil {
			zap.L().Error("Error closing Kafka reader", zap.Error(err))
		}
	}()

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err == nil {
			if ctx.Err() != nil {
				zap.L().Info("Kafka consumer is stopping (context was canceled)")
				return
			}
			zap.L().Info("New product event",
				zap.Int64("offset", msg.Offset),
				zap.Int("partition", msg.Partition),
				zap.String("key", string(msg.Key)),
				zap.String("value", string(msg.Value)),
			)
		}
	}
}

func (c *Customer) Close() error {
	return c.reader.Close()
}
