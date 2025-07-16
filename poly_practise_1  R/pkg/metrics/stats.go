package metrics

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type StatsCollector struct {
	logger    *zap.Logger
	mu        sync.RWMutex
	startTime time.Time
}

var globalStats *StatsCollector

func InitStatsCollector(logger *zap.Logger) {
	globalStats = &StatsCollector{
		logger:    logger,
		startTime: time.Now(),
	}
	lastStatsTime = time.Now()

	atomic.StoreInt64(&producerSentCounter, 0)
	atomic.StoreInt64(&consumerReceivedCounter, 0)
}

func StartStatsLoop(ctx context.Context) {
	if globalStats == nil {
		return
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			globalStats.printStats()
		}
	}
}

func (s *StatsCollector) printStats() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(s.startTime)
	now := time.Now()

	// Рассчитываем rate
	producerSent := GetProducerSent()
	consumerReceived := GetConsumerReceived()

	var producerRate, consumerRate float64
	if !lastStatsTime.IsZero() {
		timeDiff := now.Sub(lastStatsTime)
		producerRate = calculateRate(producerSent, lastProducerSent, timeDiff)
		consumerRate = calculateRate(consumerReceived, lastConsumerReceived, timeDiff)
	}

	// Обновляем для следующего расчета
	lastProducerSent = producerSent
	lastConsumerReceived = consumerReceived
	lastStatsTime = now

	lag := producerSent - consumerReceived
	MessageLag.Set(float64(lag))

	s.logger.Info("STATS",
		zap.Int64("producer_sent", producerSent),
		zap.Float64("producer_rate", producerRate),
		zap.Int64("consumer_received", consumerReceived),
		zap.Float64("consumer_rate", consumerRate),
		zap.Int64("lag", lag),
		zap.Int("system_memory_mb", int(m.Alloc)/1024/1024),
		zap.Int("system_goroutines", runtime.NumGoroutine()),
		zap.Duration("uptime", uptime),
	)
}

// Atomic счетчики для компонентов
var (
	producerSentCounter     int64
	consumerReceivedCounter int64

	// Для расчета rate
	lastProducerSent     int64
	lastConsumerReceived int64
	lastStatsTime        time.Time
)

func calculateRate(current, last int64, timeDiff time.Duration) float64 {
	if timeDiff <= 0 {
		return 0
	}
	return float64(current-last) / timeDiff.Seconds()
}

func IncrementProducerSent() {
	atomic.AddInt64(&producerSentCounter, 1)
	ProducerMessagesSent.Inc()
}

func IncrementConsumerReceived() {
	atomic.AddInt64(&consumerReceivedCounter, 1)
	ConsumerMessagesReceived.Inc()
}

func GetProducerSent() int64 {
	return atomic.LoadInt64(&producerSentCounter)
}

func GetConsumerReceived() int64 {
	return atomic.LoadInt64(&consumerReceivedCounter)
}
