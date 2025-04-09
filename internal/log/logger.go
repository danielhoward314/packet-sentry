package log

import (
	"log/slog"
	"runtime"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/danielhoward314/packet-sentry/internal/config"
)

// GetBaseLogger returns the base instance of the structured logger
func GetBaseLogger() *slog.Logger {
	logFilePath := config.GetLogFilePath()

	rotatingLog := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     28,
		Compress:   true,
	}

	handler := slog.NewJSONHandler(rotatingLog, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler)
	logger.With(KeyArch, runtime.GOARCH)
	logger.With(KeyOS, runtime.GOOS)
	slog.SetDefault(logger)
	return logger
}
