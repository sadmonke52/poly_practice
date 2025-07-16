package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"
	"myproducer/config"
	"myproducer/internal/logging"
	"myproducer/internal/producer"
	"os"
	"os/signal"
	"syscall"
)

//$env:KAFKA_BROKER = "localhost:29092"

func main() {

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/myapp/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal(err)
	}
	logger, err := logging.NewLogger(*cfg.Logger)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	logging.StatusLogger(logger, *cfg.Logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for i := 0; i < cfg.Producer.ProducerInstance; i++ {
		go func(instanceID int) {
			prod := producer.New(*cfg.Producer, cfg.Producer.Brokers, logger.With(zap.Int("instance_id", instanceID)))
			if err := prod.Run(ctx); err != nil {
				logger.Error("producer error", zap.Error(err))
			}
		}(i)
	}

	<-sigCh
	fmt.Println("Shutdown signal received")
	cancel()
}
