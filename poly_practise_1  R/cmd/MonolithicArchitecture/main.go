package main

import (
	"context"
	"log"
	"os/signal"
	"poly_practice_1/config"
	"poly_practice_1/internal/consumer"
	"poly_practice_1/internal/producer"
	"poly_practice_1/pkg/metrics"
	"syscall"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg, err := config.Load("config/local.yaml")
	if err != nil {
		log.Fatal(err)
	}
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	zapCfg := zap.NewProductionConfig()
	logger, _ := zapCfg.Build()
	zap.ReplaceGlobals(logger)

	if err := run(rootCtx, cfg); err != nil {
		logger.Fatal("service exited with error", zap.Error(err))
	}
}

func run(ctx context.Context, cfg *config.Config) error {
	grp, ctx := errgroup.WithContext(ctx)

	logger, _ := zap.NewProductionConfig().Build()
	metrics.InitStatsCollector(logger)
	go metrics.StartStatsLoop(ctx)

	redis := redis.NewClient(&redis.Options{
		Addr: cfg.RedisDB.Address, Password: cfg.RedisDB.Password, DB: cfg.RedisDB.DB,
	})
	defer redis.Close()

	for i := 0; i < cfg.Instances.ProducerCount; i++ {
		producer := producer.New(*cfg.Producer, cfg.Producer.Brokers)
		grp.Go(func() error { return producer.Run(ctx) })
	}

	for i := 0; i < cfg.Instances.ConsumerCount; i++ {
		consumer := consumer.New(*cfg.Kafka, redis)
		grp.Go(func() error { return consumer.Run(ctx) })
	}

	grp.Go(func() error { return metrics.Serve(ctx, ":8082") })

	return grp.Wait()
}
