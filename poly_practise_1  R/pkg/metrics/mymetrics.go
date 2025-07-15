package metrics

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func MonitorKafkaChannel(ctx context.Context, logger *zap.Logger, chName string, ch chan kafka.Message) {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				UpdateKafkaChannelUsage(chName, len(ch), cap(ch))
			}
		}
	}()
}
