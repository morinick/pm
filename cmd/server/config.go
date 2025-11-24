package main

import (
	"log/slog"
	"os"
)

type config struct {
	LogLevel  slog.Level
	DBURL     string
	BackupDir string
	AssetsDir string
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

	cfg.DBURL = "data.db"
	cfg.BackupDir = "/backup"
	cfg.AssetsDir = "/assets"

	cfg.MasterKey = loadMasterKey()

	return cfg, nil
}

func loadMasterKey() string {
	key, _ := os.ReadFile("master.key")
	if len(key) == 0 {
		return os.Getenv("MASTER_KEY")
	}
	return string(key)
}

func saveMasterKey(key string) error {
	keyFile, err := os.Create("master.key")
	if err != nil {
		return err
	}
	_, err = keyFile.WriteString(key)
	return err
}
