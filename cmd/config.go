package main

import (
	"log/slog"
	"os"
)

type config struct {
	LogLevel slog.Level
}

var logLevelMap = map[string]slog.Level{
	"DEBUG": slog.LevelDebug,
	"INFO":  slog.LevelInfo,
	"WARN":  slog.LevelWarn,
	"ERROR": slog.LevelError,
}

func newConfig() (config, error) {
	cfg := config{}

	var ok bool
	if cfg.LogLevel, ok = logLevelMap[os.Getenv("LOG_LEVEL")]; !ok {
		cfg.LogLevel = slog.LevelInfo
	}

	return cfg, nil
}
