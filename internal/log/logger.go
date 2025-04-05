package log

import (
	"log/slog"
	"runtime"

	"gopkg.in/natefinch/lumberjack.v2"
)

func getLogFilePath() string {
	if runtime.GOOS == "windows" {
		return `C:\Program Files\PacketSentry\packet-sentry-agent.log`
	}
	return "/var/log/packet-sentry-agent.log"
}

func GetBaseLogger() *slog.Logger {
	logFilePath := getLogFilePath()

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
