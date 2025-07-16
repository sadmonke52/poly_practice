package app

import (
	"collector/config"
	"collector/internal/aggregator"
	"collector/internal/consumer"
	"collector/internal/flusher"
	"collector/internal/producer"
	"collector/pkg/logging"
	"collector/pkg/mymetrics"
	"context"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func MustRun(cfg *config.Config) {
	logger, err := logging.NewLogger(*cfg.Logger)
	if err != nil {
		panic(err)
	}
	logging.StatusLogger(logger, *cfg.Logger)

	rand.Seed(time.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go serveMetrics(logger)

	agg := aggregator.New()

	cons, err := consumer.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to init consumer", zap.Error(err))
	}
	defer func() {
		if err := cons.Close(); err != nil {
			logger.Error("failed to close consumer", zap.Error(err))
		}
	}()
	msgCh := cons.StartConsuming(ctx)
	go aggregator.StartAggregatorLoop(ctx, msgCh, agg)

	prod, err := producer.New(cfg.Producer, logger)
	if err != nil {
		logger.Fatal("failed to init producer", zap.Error(err))
	}
	defer func() {
		if err := prod.Close(); err != nil {
			logger.Error("failed to close producer", zap.Error(err))
		}
	}()

	for _, topic := range cfg.Producer.Topics {
		flush := flusher.New(agg, logger, prod, cfg.Producer.FlushSec, topic)
		go flush.Run(ctx)
	}

	waitForSignal(logger)
	cancel()
	logger.Info("Shutdown complete")
}

func serveMetrics(logger *zap.Logger) {
	mymetrics.Init()
	http.Handle("/metrics", mymetrics.Handler())
	if err := http.ListenAndServe(":9200", nil); err != nil {
		logger.Fatal("metrics server error", zap.Error(err))
	}
}

func waitForSignal(logger *zap.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Info("shutdown signal received", zap.String("signal", sig.String()))
}
