package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"processor/config"
	"strings"
)

func NewLogger(cfg config.LoggerConfig) (*zap.Logger, error) {

	var level zapcore.Level
	err := level.UnmarshalText([]byte(strings.ToLower(cfg.Level)))
	if err != nil {
		level = zapcore.InfoLevel
	}

	var encoderCfg zapcore.EncoderConfig
	if strings.ToLower(cfg.Format) == "console" {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderCfg = zap.NewProductionEncoderConfig()
	}

	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if strings.ToLower(cfg.Format) == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	var outputWriter zapcore.WriteSyncer
	switch strings.ToLower(cfg.Output) {
	case "stderr":
		outputWriter = zapcore.Lock(os.Stderr)
	case "stdout":
		outputWriter = zapcore.Lock(os.Stdout)
	}

	core := zapcore.NewCore(encoder, outputWriter, level)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return logger, nil
}

func StatusLogger(logger *zap.Logger, cfg config.LoggerConfig) {
	logger.Info("logger initialized",
		zap.String("level", cfg.Level),
		zap.String("format", cfg.Format),
		zap.String("output", cfg.Output),
	)
}
