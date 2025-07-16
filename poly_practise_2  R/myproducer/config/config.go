package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Producer *ProducerConfig `yaml:"producer"`
	Logger   *LoggerConfig   `yaml:"logging"`
}
type ProducerConfig struct {
	Brokers          []string `yaml:"brokers" env-required:"true"`
	Topics           []string `yaml:"topics"` // для нескольких топиков
	UserCount        int      `yaml:"usercount" env-default:"1"`
	MessageCount     int      `yaml:"messagecount" env-default:"1"`
	Throughput       int      `yaml:"throughput" env-default:"1"`
	Workers          int      `yaml:"workers" env-default:"1"` // Worker pool size
	ProducerInstance int      `yaml:"producer-instance" env-default:"1"`
}

type LoggerConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output" env-default:"stdout"`
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
