package main

import (
	"log/slog"
	"os"
)

type config struct {
	LogLevel  slog.Level
	MasterKey string
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

	cfg.MasterKey = loadMasterKey()

	return cfg, nil
}

func loadMasterKey() string {
	key, _ := os.ReadFile("master_key")
	if len(key) == 0 {
		return os.Getenv("MASTER_KEY")
	}
	return string(key)
}

func saveMasterKey(key string) error {
	keyFile, err := os.Create("master_key")
	if err != nil {
		return err
	}
	_, err = keyFile.WriteString(key)
	return err
}
