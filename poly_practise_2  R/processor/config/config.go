package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	HTTPServer *HttpServer       `yaml:"http-server"`
	Logger     *LoggerConfig     `yaml:"logging"`
	Kafka      *KafkaConfig      `yaml:"kafka"`
	Aggregator *AggregatorConfig `yaml:"aggregator"`
	Redis      *RedisConfig      `yaml:"redis"`
}

type HttpServer struct {
	Address     string        `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type LoggerConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output" env-default:"stdout"`
}

type KafkaConfig struct {
	Brokers           []string `yaml:"brokers" env-required:"true"`
	Topics            []string `yaml:"topics" env-required:"true"`
	GroupID           string   `yaml:"groupid" env-required:"true"`
	ClientID          string   `yaml:"clientid"`
	WorkerCount       int      `yaml:"worker-count" env-default:"1"`
	ReaderInstance    int      `yaml:"reader-instance" env-default:"2"`
	BufferChannelSize int      `yaml:"buffer-channel-size" env-default:"100"`
}

type AggregatorConfig struct {
	AggregationWindow time.Duration `yaml:"aggregationwindow" env-default:"5s"` // Как часто сбрасывать агрегированные данные
}
type RedisConfig struct {
	Addr     string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if err := envconfig.Process("", &cfg); err != nil { // для переменных окруженияя
		return nil, fmt.Errorf("read env: %w", err)
	}

	return &cfg, nil
}
