package producer

import (
	"collector/config"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"math/rand"
	"sync/atomic"
)

type Producer struct {
	cfg      *config.ProducerConfig
	writers  []*kafka.Writer
	log      *zap.Logger
	counters map[string]*int64
}

func New(cfg *config.ProducerConfig, logger *zap.Logger) (*Producer, error) {
	ws := make([]*kafka.Writer, len(cfg.Topics))
	counters := make(map[string]*int64)
	for i, t := range cfg.Topics {
		ws[i] = &kafka.Writer{
			Addr:     kafka.TCP(cfg.Brokers...),
			Topic:    t,
			Balancer: &kafka.LeastBytes{},
		}

		var zero int64
		counters[t] = &zero
	}
	return &Producer{
		cfg:      cfg,
		writers:  ws,
		log:      logger,
		counters: counters,
	}, nil
}

func (p *Producer) Send(ctx context.Context, userID string, items []string) error {

	topicIdx := rand.Intn(len(p.writers))

	value, err := json.Marshal(items)
	if err != nil {
		p.log.Error("failed to marshal items", zap.String("userID", userID), zap.Error(err))
		return err
	}

	msg := kafka.Message{
		Key:   []byte(userID),
		Value: value,
	}

	err = p.writers[topicIdx].WriteMessages(ctx, msg)
	if err != nil {
		p.log.Error("failed to write message",
			zap.String("userID", userID),
			zap.String("topic", p.writers[topicIdx].Topic),
			zap.Error(err),
		)
		return err
	}

	count := atomic.AddInt64(p.counters[p.writers[topicIdx].Topic], 1)
	if count%100 == 0 {
		p.log.Info("sent 100 messages", zap.String("topic", p.writers[topicIdx].Topic), zap.Int64("total_sent", count))
	}
	return nil
}

func (p *Producer) Close() error {
	for _, w := range p.writers {
		if err := w.Close(); err != nil {
			p.log.Warn("failed to close kafka writer", zap.String("topic", w.Topic), zap.Error(err))
		}
	}
	return nil
}
