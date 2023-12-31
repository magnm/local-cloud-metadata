package server

import (
	"os"

	"github.com/magnm/lcm/config"
	"golang.org/x/exp/slog"
)

func setupLogging(cfg config.Config) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: configLevelToSlogLevel(cfg.LogLevel),
	}))
	slog.SetDefault(logger)
	slog.Info("logging initialised", "level", cfg.LogLevel)
}

func configLevelToSlogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
