package logger

import (
	"log/slog"
	"os"
)

type Config struct{
	Level string `default:"warn" envconfig:"LOGGER_LEVEL"`
}

var logger *slog.Logger

func Init(c Config){
	var level slog.Level

	switch c.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelWarn
	}

	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Error(msg string, err error, args ...any) {
	logger.Error(msg, append(args, slog.Any("error", err))...)
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}