package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Kafka      *KafkaConfig      `yaml:"kafka"`
	RedisDB    *RedisСonfig      `yaml:"redis_database"`
	Logging    *LoggingConfig    `yaml:"logging"`
	Producer   *ProducerConfig   `yaml:"producer"`
	Aggregator *AggregatorConfig `yaml:"aggregator"`
	Instances  *InstancesConfig  `yaml:"instances"`
}

type KafkaConfig struct {
	Broker      []string `yaml:"broker" env-required:"true"` // FIX
	Brokers     []string `yaml:"brokers"`
	Topic       string   `yaml:"topic" env-required:"true"`
	Topics      []string `yaml:"topics" env-required:"true"`
	GroupID     string   `yaml:"group-id" env-required:"true"`
	ClientID    string   `yaml:"client_id"`
	WorkerCount int      `yaml:"worker-count" env-default:"1"` //FIX
	ReaderCount int      `yaml:"reader-count" env-default:"3"`
	Workers     int      `yaml:"workers" env-default:"1"` // Worker pool size
}

type Kafka struct {
	Brokers []string `yaml:"brokers" env-required:"true"`
	GroupID string   `yaml:"group-id" env-required:"true"`
	Topic   string   `yaml:"topic" env-required:"true"`
	Workers int      `yaml:"workers" env-default:"1"`
}

type AggregatorConfig struct {
	FlushInterval time.Duration `yaml:"flush-interval" env-default:"5s"`
	BatchSize     int           `yaml:"batch-size" env-default:"100"`
}

type RedisСonfig struct {
	Address  string `yaml:"address" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	DB       int    `yaml:"db" env-required:"true"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output" env-default:"stdout"`
}
type ProducerConfig struct {
	Brokers      []string `yaml:"brokers" env-required:"true"`
	Topic        string   `yaml:"topic" env-required:"true"`
	Topics       []string `yaml:"topics"` // для нескольких топиков
	UserCount    int      `yaml:"usercount" env-default:"1"`
	MessageCount int      `yaml:"messagecount" env-default:"1"`
	Throughput   int      `yaml:"throughput" env-default:"1"`
	Workers      int      `yaml:"workers" env-default:"1"` // Worker pool size
}

type InstancesConfig struct {
	ProducerCount int `yaml:"producer_count" env-default:"1"`
	ConsumerCount int `yaml:"consumer_count" env-default:"1"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
