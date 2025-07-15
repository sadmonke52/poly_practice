package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"processor/config"
	"processor/internal/stress-tester"
	"processor/pkg/metrics"
)

type Aggregator struct {
	cfg             *config.AggregatorConfig
	logger          *zap.Logger
	inputChan       <-chan kafka.Message
	aggregatorState map[string]int64
	redisClient     *redis.Client
	redisCfg        *config.RedisConfig
}

func New(cfg *config.AggregatorConfig, logger *zap.Logger, inputChan <-chan kafka.Message, redisCfg *config.RedisConfig,
) (*Aggregator, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})

	return &Aggregator{
		cfg:             cfg,
		logger:          logger,
		inputChan:       inputChan,
		aggregatorState: make(map[string]int64),
		redisClient:     rdb,
		redisCfg:        redisCfg,
	}, nil
}

func (a *Aggregator) StartProcessing(ctx context.Context) {
	a.logger.Info("Central aggregator processing loop started",
		zap.Duration("aggregation_window", a.cfg.AggregationWindow),
	)

	metrics.QueueSize.WithLabelValues("aggregator_input").Set(float64(len(a.inputChan)))

	ticker := time.NewTicker(a.cfg.AggregationWindow)
	defer ticker.Stop()

	defer func() {
		if err := a.redisClient.Close(); err != nil {
			a.logger.Error("failed to close Redis client", zap.Error(err))
		}
	}()

	for {
		select {
		case msg, ok := <-a.inputChan:
			metrics.QueueSize.WithLabelValues("aggregator_input").Dec()
			if !ok {
				a.logger.Info("Input channel closed, flushing final data...")
				a.flushResults()
				return
			}

			_ = processMessageValue(msg.Value, a.logger)

			metrics.MessagesConsumed.Inc()
			metrics.InFlightMessages.Inc()

			key := string(msg.Key)
			stress_tester.SimulateHeavyGCPollution()
			a.aggregatorState[key]++

			metrics.InFlightMessages.Dec()

		case <-ticker.C:
			a.flushResults()
			a.aggregatorState = make(map[string]int64)

		case <-ctx.Done():
			a.logger.Info("Context cancelled, flushing final data...")
			a.flushResults()
			return
		}
	}
}

func (a *Aggregator) flushResults() {
	if len(a.aggregatorState) == 0 {
		return
	}

	a.logger.Info("Flushing aggregated data to Redis",
		zap.Int("unique_keys_count", len(a.aggregatorState)),
	)

	redisHashKey := fmt.Sprintf("agg_batch:%s", time.Now().Format("2006-01-02T15:04:05"))

	fields := make(map[string]interface{}, len(a.aggregatorState))
	for k, v := range a.aggregatorState {
		fields[k] = v
	}

	if err := a.redisClient.HMSet(a.redisClient.Context(), redisHashKey, fields).Err(); err != nil {
		a.logger.Error("failed to write aggregated data to Redis", zap.Error(err), zap.String("redis_key", redisHashKey))
		metrics.MessagesFailed.WithLabelValues("redis_write_failed").Inc()
	} else {
		a.logger.Info("Successfully wrote aggregated data to Redis", zap.String("redis_key", redisHashKey), zap.Int("fields_count", len(fields)))
	}
}

func processMessageValue(value []byte, logger *zap.Logger) error {
	var records []string
	if err := json.Unmarshal(value, &records); err != nil {
		logger.Error("failed to unmarshal Kafka message value", zap.Error(err))
		return err
	}

	for _, r := range records {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(r), &obj); err != nil {
			logger.Warn("skipping invalid record", zap.Error(err))
		}
	}
	return nil
}
