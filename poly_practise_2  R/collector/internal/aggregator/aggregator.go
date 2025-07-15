package aggregator

import (
	"collector/pkg/mymetrics"
	"context"
	"github.com/segmentio/kafka-go"
	"sync"
)

type Aggregator struct {
	mu    sync.RWMutex
	batch map[string][]string
}

func New() *Aggregator {
	return &Aggregator{
		batch: make(map[string][]string),
	}
}

func StartAggregatorLoop(ctx context.Context, msgCh <-chan kafka.Message, agg *Aggregator,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgCh:
			if !ok {
				return
			}

			userID := getHeader(msg, "auth_user_id")
			if userID == "" {
				continue
			}
			agg.Add(msg.Topic, userID, string(msg.Value))
		}
	}
}

func (a *Aggregator) Add(topic, userID, item string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.batch[userID] = append(a.batch[userID], item)
	mymetrics.InFlightMessages.WithLabelValues(topic).Inc()
}
func countMessages(data map[string][]string) float64 {
	var total float64
	for _, items := range data {
		total += float64(len(items))
	}
	return total
}
func (a *Aggregator) DrainAndReset(topic string) map[string][]string {
	a.mu.Lock()
	defer a.mu.Unlock()

	data := a.batch
	a.batch = make(map[string][]string)

	totalMessages := countMessages(data)

	mymetrics.InFlightMessages.WithLabelValues(topic).Add(-totalMessages)
	return data
}

func getHeader(msg kafka.Message, key string) string {
	for _, h := range msg.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}
