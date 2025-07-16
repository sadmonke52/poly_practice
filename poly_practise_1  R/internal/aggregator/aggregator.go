package aggregator

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Aggregator struct {
	redis *redis.Client
	batch map[string][]string
	mu    sync.Mutex
}

func New(rdb *redis.Client) *Aggregator {
	return &Aggregator{redis: rdb, batch: map[string][]string{}}
}

func (a *Aggregator) Run(ctx context.Context, in <-chan kafka.Message) error {
	flush := time.NewTicker(5 * time.Second)
	defer flush.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-flush.C:
			a.flush(ctx)
		case m, ok := <-in:
			if !ok {
				a.flush(ctx)
				return nil
			}
			a.append(m)
		}
	}
}

func (a *Aggregator) append(m kafka.Message) {
	uid := header(m, "auth_user_id")
	if uid == "" {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.batch[uid] = append(a.batch[uid], string(m.Key))
}

func (a *Aggregator) flush(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()

	pipe := a.redis.Pipeline()
	for uid, items := range a.batch {
		data, _ := json.Marshal(items)
		pipe.Set(ctx, "agg:"+uid, data, 0)
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, context.Canceled) {
		zap.L().Error("redis pipeline", zap.Error(err))
	}
	a.batch = map[string][]string{}
}

func header(m kafka.Message, key string) string {
	for _, h := range m.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (a *Aggregator) Consume(ctx context.Context, msgCh <-chan kafka.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgCh:
			if !ok {
				return
			}
			a.append(msg)
		}
	}
}

func (a *Aggregator) FlushLoop(ctx context.Context, cfg interface{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.flush(ctx)
		}
	}
}

func (a *Aggregator) GetData() map[string][]string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.batch
}

func (a *Aggregator) GetUsersCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.batch)
}
