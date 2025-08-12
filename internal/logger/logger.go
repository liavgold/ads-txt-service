package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

type Logger struct {
	*zap.SugaredLogger
}

var (
	instance *Logger
	once     sync.Once
)

func Init(level string) error {
	var initErr error
	once.Do(func() {
		zapLevel := parseLevel(level)

		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapLevel)
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		zapLogger, err := cfg.Build()
		if err != nil {
			initErr = err
			return
		}

		instance = &Logger{SugaredLogger: zapLogger.Sugar()}
	})

	return initErr
}

func L() *Logger {
	if instance == nil {
		panic("logger not initialized, call logger.Init first")
	}
	return instance
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
