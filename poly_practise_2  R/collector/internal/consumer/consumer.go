package consumer

import (
	"collector/config"
	"context"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
)

type Consumer struct {
	cfg     *config.KafkaConfig
	readers map[string][]*kafka.Reader // много ридеров не хорошо наверное лучше запускать консюмеры
	logger  *zap.Logger
}

func New(cfg *config.Config, logger *zap.Logger) (*Consumer, error) {
	cons := &Consumer{
		cfg:     cfg.Kafka,
		readers: make(map[string][]*kafka.Reader),
		logger:  logger,
	}

	for _, topic := range cfg.Kafka.Topics {
		for i := 0; i < cfg.Kafka.ReaderInstance; i++ {
			r := kafka.NewReader(kafka.ReaderConfig{
				Brokers:  cfg.Kafka.Brokers,
				GroupID:  cfg.Kafka.GroupID,
				Topic:    topic,
				MinBytes: 10e3,
				MaxBytes: 10e6,
			})
			cons.readers[topic] = append(cons.readers[topic], r)
		}
	}
	return cons, nil
}

func (c *Consumer) StartConsuming(ctx context.Context) <-chan kafka.Message {
	msgCh := make(chan kafka.Message, c.cfg.BufferChannelSize)

	var wg sync.WaitGroup
	var messagesRead uint64

	for topic, readers := range c.readers {
		for _, r := range readers {
			wg.Add(1)
			go func(topic string, rd *kafka.Reader) {
				defer wg.Done()
				for {
					msg, err := rd.ReadMessage(ctx)
					if err != nil {
						if err == context.Canceled || ctx.Err() != nil {
							c.logger.Info("stopping consumer for topic", zap.String("topic", topic))
							return
						}
						c.logger.Error("failed to read message", zap.String("topic", topic), zap.Error(err))
						continue
					}

					if n := atomic.AddUint64(&messagesRead, 1); n%100 == 0 {
						c.logger.Info("consumer is active",
							zap.Uint64("messages_read_total", n),
							zap.String("topic", topic),
						)
					}

					select {
					case msgCh <- msg:
					case <-ctx.Done():
						return
					}
				}
			}(topic, r)
		}
	}

	go func() {
		wg.Wait()
		close(msgCh)
	}()
	return msgCh
}

func (c *Consumer) Close() error {
	for topic, readers := range c.readers {
		for _, r := range readers {
			if err := r.Close(); err != nil {
				c.logger.Error("failed to close kafka reader", zap.String("topic", topic), zap.Error(err))
			} else {
				c.logger.Info("kafka reader closed", zap.String("topic", topic))
			}
		}
	}
	return nil
}
