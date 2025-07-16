package app

import (
	"context"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"processor/config"
	"processor/internal/aggregator"
	"processor/internal/consumer"
	"processor/pkg/logging"
	"processor/pkg/metrics"
	"sync"
	"syscall"
)

func MustRun(cfg *config.Config) {
	logger, err := logging.NewLogger(*cfg.Logger)
	if err != nil {
		panic(err)
	}
	logging.StatusLogger(logger, *cfg.Logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	go serveMetrics()

	consumers := make([]*consumer.Consumer, 0, cfg.Kafka.ReaderInstance)

	mergedChan := make(chan kafka.Message, cfg.Kafka.BufferChannelSize)
	var producerWg sync.WaitGroup

	logger.Info("Starting consumer instances", zap.Int("n", cfg.Kafka.ReaderInstance))
	for i := 0; i < cfg.Kafka.ReaderInstance; i++ {
		cons, err := consumer.New(cfg, logger)
		if err != nil {
			logger.Fatal("failed to init consumer", zap.Error(err))
		}
		consumers = append(consumers, cons)

		inputChan := cons.StartConsuming(ctx)

		producerWg.Add(1)
		wg.Add(1)
		go func(instance int) {
			defer producerWg.Done()
			defer wg.Done()

			for {
				select {
				case msg, ok := <-inputChan:
					if !ok {
						logger.Info("Input channel closed, forwarder exiting", zap.Int("instance", instance))
						return
					}
					select {
					case mergedChan <- msg:
					case <-ctx.Done():
						logger.Info("Shutdown signal received while forwarding message, exiting", zap.Int("instance", instance))
						return
					}
				case <-ctx.Done():
					logger.Info("Shutdown signal received, forwarder exiting", zap.Int("instance", instance))
					return
				}
			}
		}(i)
	}

	go func() {
		producerWg.Wait()
		close(mergedChan)
		logger.Info("All producers finished, closing merged channel.")
	}()

	agg, err := aggregator.New(cfg.Aggregator, logger, mergedChan, cfg.Redis)
	if err != nil {
		logger.Fatal("failed to init aggregator", zap.Error(err))
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		agg.StartProcessing(ctx)
		logger.Info("Aggregator has finished processing.")
	}()

	waitForSignal(logger)
	logger.Info("Shutdown signal received. Shutting down...")

	cancel()
	wg.Wait()
	logger.Info("All worker goroutines have been shut down.")

	for i, c := range consumers {
		if err := c.Close(); err != nil {
			logger.Error("failed to close consumer", zap.Error(err), zap.Int("instance", i))
		}
	}

	logger.Info("Shutdown complete")
}

func serveMetrics() {
	metrics.Init()
	http.Handle("/metrics", metrics.Handler())
	_ = http.ListenAndServe(":9200", nil)
}

func waitForSignal(logger *zap.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Info("shutdown signal received", zap.String("signal", sig.String()))
}
