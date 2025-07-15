package consumer

import (
	"context"
	"fmt"
	"math/rand"
	"poly_practice_1/config"
	"poly_practice_1/internal/aggregator"
	"poly_practice_1/pkg/metrics"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Consumer struct {
	kafka   *kafka.Reader
	agg     *aggregator.Aggregator
	workers int
}

func New(cfg config.KafkaConfig, redis *redis.Client) *Consumer {

	workers := cfg.Workers

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    cfg.Topic,
		MinBytes: 10e3, MaxBytes: 10e6,
	})
	return &Consumer{
		kafka:   r,
		agg:     aggregator.New(redis),
		workers: workers,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	defer c.kafka.Close()

	msgCh := make(chan kafka.Message, 1000000) // !!!
	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		defer close(msgCh)
		for {
			m, err := c.kafka.ReadMessage(ctx)
			if err != nil {
				return err
			}
			select {
			case msgCh <- m:
				metrics.IncrementConsumerReceived()
				simulateHeavyGCPollution()
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	for i := 0; i < c.workers; i++ {
		wg.Go(func() error { return c.agg.Run(ctx, msgCh) })
	}
	return wg.Wait()
}

func simulateHeavyGCPollution() {
	size := 1<<20 + rand.Intn(3<<20)
	_ = make([]byte, size)

	for i := 0; i < 1000; i++ {
		_ = fmt.Sprintf("%d-%d-%d", rand.Int63(), rand.Int63(), rand.Int63())
	}

	m := make(map[string][]byte, 1000)
	for i := 0; i < 10000; i++ {
		m[strconv.Itoa(i)] = make([]byte, 8<<10)
	}
}

func (c Consumer) RunWithConfig(
	ctx context.Context,
	cfg *config.Config,
	logger *zap.Logger,
	agg *aggregator.Aggregator,
) error {
	if c.agg != nil {
		return c.Run(ctx)
	}

	msgCh := make(chan kafka.Message, 1000000)

	metrics.MonitorKafkaChannel(ctx, logger, "Consumer", msgCh)

	go func() {
		agg.Consume(ctx, msgCh)
	}()

	go func() {
		agg.FlushLoop(ctx, cfg.Aggregator)
	}()

	for {
		msg, err := c.kafka.ReadMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				return ctx.Err()
			}
			logger.Error("Error reading message", zap.Error(err))
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case msgCh <- msg:
			metrics.IncrementConsumerReceived()
		}
	}
}
