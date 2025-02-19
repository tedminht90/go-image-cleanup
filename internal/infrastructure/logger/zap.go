package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Level      string
	LogDir     string
	MaxSize    int  // megabytes
	MaxBackups int  // number of backups
	MaxAge     int  // days
	Compress   bool // compress old files
}

func NewLogger(cfg Config) (*zap.Logger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("can't create log directory: %w", err)
	}

	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("can't parse log level: %w", err)
	}

	// Create service log writer with rotation
	serviceLogWriter := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.LogDir, "service.log"),
		MaxSize:    cfg.MaxSize,    // megabytes
		MaxBackups: cfg.MaxBackups, // number of backups
		MaxAge:     cfg.MaxAge,     // days
		Compress:   cfg.Compress,   // compress old files
	}

	// Create error log writer with rotation
	errorLogWriter := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.LogDir, "error.log"),
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	// Create encoders
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create cores
	serviceCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(serviceLogWriter),
		level,
	)

	errorCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(errorLogWriter),
		zapcore.ErrorLevel,
	)

	// Create logger
	logger := zap.New(
		zapcore.NewTee(serviceCore, errorCore),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return logger, nil
}
