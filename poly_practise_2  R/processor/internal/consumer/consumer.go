package consumer

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"processor/config"
	"sync"
	"sync/atomic"
)

type Consumer struct {
	cfg     *config.KafkaConfig
	readers map[string]*kafka.Reader
	logger  *zap.Logger
}

func New(cfg *config.Config, logger *zap.Logger) (*Consumer, error) {
	cons := &Consumer{
		cfg:     cfg.Kafka,
		readers: make(map[string]*kafka.Reader),
		logger:  logger,
	}

	for _, topic := range cfg.Kafka.Topics {
		cons.readers[topic] = kafka.NewReader(kafka.ReaderConfig{
			Brokers:  cfg.Kafka.Brokers,
			GroupID:  cfg.Kafka.GroupID,
			Topic:    topic,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		})
	}
	return cons, nil
}

func (c *Consumer) StartConsuming(ctx context.Context) <-chan kafka.Message {
	msgCh := make(chan kafka.Message, c.cfg.BufferChannelSize)
	var wg sync.WaitGroup

	var messagesRead uint64 = 0

	for topic, reader := range c.readers {
		wg.Add(1)
		go func(topic string, r *kafka.Reader) {
			defer wg.Done()
			c.logger.Info("Starting consumer loop for topic", zap.String("topic", topic))

			for {

				c.logger.Debug("Attempting to read message from Kafka", zap.String("topic", topic))

				msg, err := r.ReadMessage(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						c.logger.Info("Context cancelled, stopping consumer", zap.String("topic", topic))
						return
					}
					c.logger.Error("Failed to read message from Kafka", zap.String("topic", topic), zap.Error(err))
					if ctx.Err() != nil {
						return
					}
					continue
				}

				c.logger.Debug("Message read successfully", zap.String("topic", topic), zap.Int64("offset", msg.Offset))

				atomic.AddUint64(&messagesRead, 1)

				select {
				case msgCh <- msg:
				case <-ctx.Done():
					c.logger.Info("Context cancelled while sending message to channel", zap.String("topic", topic))
					return
				}
			}
		}(topic, reader)
	}

	go func() {
		wg.Wait()
		close(msgCh)
		c.logger.Info("All topic consumers have shut down, closing messages channel.")
	}()

	return msgCh
}

func (c *Consumer) Close() error {
	for topic, r := range c.readers {
		if err := r.Close(); err != nil {
			c.logger.Error("failed to close kafka reader", zap.String("topic", topic), zap.Error(err))
		} else {
			c.logger.Info("kafka reader closed", zap.String("topic", topic))
		}
	}
	return nil
}
