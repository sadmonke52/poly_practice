package flusher

import (
	"collector/internal/aggregator"
	"collector/pkg/mymetrics"
	"context"
	"go.uber.org/zap"
	"time"
)

// ..
type Sender interface {
	Send(ctx context.Context, userID string, items []string) error
}

type Flusher struct {
	agg       *aggregator.Aggregator
	log       *zap.Logger
	sender    Sender
	interval  time.Duration
	topicName string
}

func New(agg *aggregator.Aggregator, log *zap.Logger, sender Sender,
	interval time.Duration, topic string) *Flusher {

	return &Flusher{
		agg:       agg,
		log:       log,
		sender:    sender,
		interval:  interval,
		topicName: topic,
	}
}

func countMessagesInBatch(data map[string][]string) float64 {
	var total float64
	for _, items := range data {
		total += float64(len(items))
	}
	return total
}

func (f *Flusher) Run(ctx context.Context) {
	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.log.Info("Flusher context cancelled, flushing before exit")
			f.flush(ctx)
			return
		case <-ticker.C:
			f.flush(ctx)
		}
	}
}

func (f *Flusher) flush(ctx context.Context) {
	data := f.agg.DrainAndReset(f.topicName)
	totalMessagesInBatch := countMessagesInBatch(data)
	totalUsers := len(data)

	f.log.Info("Batch flushed",
		zap.String("topic", f.topicName),
		zap.Int("user_count", totalUsers),
		zap.Float64("total_messages", totalMessagesInBatch),
	)

	mymetrics.QueueSize.WithLabelValues("aggregated_batch").Set(totalMessagesInBatch)

	for uid, items := range data {
		if err := f.sender.Send(ctx, uid, items); err != nil {
			f.log.Error("send failed", zap.String("uid", uid), zap.Error(err))
			mymetrics.MessagesFailed.WithLabelValues(f.topicName, "send_error").Add(float64(len(items)))
			continue
		}
		mymetrics.MessagesConsumed.WithLabelValues(f.topicName).Add(float64(len(items)))
	}
}
