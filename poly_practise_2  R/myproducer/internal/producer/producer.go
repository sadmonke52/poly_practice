package producer

import (
	"context"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"myproducer/config"
	"myproducer/internal/generator"
	"sync"
	"sync/atomic"
	"time"
)

type Producer struct {
	cfg      config.ProducerConfig
	writers  []*kafka.Writer
	gen      generator.Generator
	logger   *zap.Logger
	counters map[string]*int64
}

func New(cfg config.ProducerConfig, brokers []string, logger *zap.Logger) *Producer {
	ws := make([]*kafka.Writer, len(cfg.Topics))
	counters := make(map[string]*int64)
	for i, t := range cfg.Topics {
		ws[i] = &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    t,
			Balancer: &kafka.LeastBytes{},
		}
		var zero int64
		counters[t] = &zero
	}
	return &Producer{
		cfg:      cfg,
		writers:  ws,
		gen:      generator.New(&cfg),
		logger:   logger,
		counters: counters,
	}
}

func (p *Producer) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	ticker := time.NewTicker(time.Second / time.Duration(p.cfg.Throughput))
	defer ticker.Stop()

	limiter := make(chan struct{}, p.cfg.Workers)

	for i := 0; i < p.cfg.MessageCount; i++ {
		select {
		case <-ctx.Done():
			p.logger.Warn("Producer stopped by context")
			wg.Wait()
			return ctx.Err()
		case <-ticker.C:
			limiter <- struct{}{}
			wg.Add(1)
			go func(id int) {
				defer func() {
					<-limiter
					wg.Done()
				}()

				topicIdx := id % len(p.writers)
				msg := p.gen.Event()
				err := p.writers[topicIdx].WriteMessages(ctx, msg)
				if err != nil {
					p.logger.Warn("Kafka write failed",
						zap.String("topic", p.writers[topicIdx].Topic),
						zap.Error(err),
					)
				}
				count := atomic.AddInt64(p.counters[p.writers[topicIdx].Topic], 1)
				if count%100 == 0 {
					p.logger.Info("sent 100 messages", zap.String("topic", p.writers[topicIdx].Topic), zap.Int64("total_sent", count))
				}

			}(i)
		}
	}

	wg.Wait()
	p.logger.Info("Producer finished sending all messages")
	return nil
}
