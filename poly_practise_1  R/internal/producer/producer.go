package producer

import (
	"context"
	"go.uber.org/zap"
	"poly_practice_1/config"
	"poly_practice_1/internal/generator"
	"poly_practice_1/pkg/metrics"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	cfg     config.ProducerConfig
	writers []*kafka.Writer
	gen     generator.Generator
}

func New(cfg config.ProducerConfig, brokers []string) *Producer {
	ws := make([]*kafka.Writer, len(cfg.Topics))
	for i, t := range cfg.Topics {
		ws[i] = &kafka.Writer{Addr: kafka.TCP(brokers...), Topic: t, Balancer: &kafka.LeastBytes{}}
	}
	return &Producer{cfg: cfg, writers: ws, gen: generator.New(&cfg)}
}

func (p *Producer) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second / time.Duration(p.cfg.Throughput))
	defer ticker.Stop()

	limiter := make(chan struct{}, p.cfg.Workers)

	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			limiter <- struct{}{}
			go func(topicIdx, id int) {
				defer func() { <-limiter }()
				msg := p.gen.Event()
				err := p.writers[topicIdx].WriteMessages(ctx, msg)
				if err != nil {
					zap.L().Warn("kafka write failed", zap.Error(err))
					return
				}
				metrics.IncrementProducerSent()
			}(i%len(p.writers), i)
		}
	}
}
