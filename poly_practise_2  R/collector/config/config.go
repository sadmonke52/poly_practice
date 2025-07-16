package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	HTTPServer *HttpServer     `yaml:"http-server"`
	Logger     *LoggerConfig   `yaml:"logging"`
	Kafka      *KafkaConfig    `yaml:"kafka"`
	Producer   *ProducerConfig `yaml:"producer"`
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
	ReaderInstance    int      `yaml:"reader-instance" env-default:"1"`
	BufferChannelSize int      `yaml:"buffer-channel-size" env-default:"100"`
}
type ProducerConfig struct {
	Brokers          []string      `yaml:"brokers" env-required:"true"`
	Topics           []string      `yaml:"topics"`
	UserCount        int           `yaml:"usercount" env-default:"1"`
	MessageCount     int           `yaml:"messagecount" env-default:"1"`
	Throughput       int           `yaml:"throughput" env-default:"1"`
	Workers          int           `yaml:"workers" env-default:"1"` // Worker pool size
	ProducerInstance int           `yaml:"producer-instance" env-default:"1"`
	FlushSec         time.Duration `yaml:"interval" env-default:"4s"`
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
