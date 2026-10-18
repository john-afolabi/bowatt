package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

type Config struct {
	ListenAddr      string
	DataDir         string
	DBPath          string
	LogLevel        slog.Level
	LogFormat       LogFormat
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	IdleTimeout     time.Duration
}

func (c Config) UploadsDir() string {
	return filepath.Join(c.DataDir, "uploads")
}

func Load() (Config, error) {
	logFormat, err := parseLogFormat(env("LOG_FORMAT", string(LogFormatText)))
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		ListenAddr:      env("LISTEN_ADDR", ":8080"),
		DataDir:         env("DATA_DIR", "./data"),
		LogLevel:        parseLevel(env("LOG_LEVEL", "info")),
		LogFormat:       logFormat,
		ShutdownTimeout: 10 * time.Second,
		ReadTimeout:     60 * time.Second,
		IdleTimeout:     120 * time.Second,
	}

	cfg.DBPath = env("DB_PATH", filepath.Join(cfg.DataDir, "app.db"))

	if cfg.ListenAddr == "" {
		return Config{}, fmt.Errorf("LISTEN_ADDR must not be empty")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(v)
	}
	return fallback
}

func parseLogFormat(value string) (LogFormat, error) {
	switch strings.ToLower(value) {
	case "", string(LogFormatText):
		return LogFormatText, nil
	case string(LogFormatJSON):
		return LogFormatJSON, nil
	default:
		return "", fmt.Errorf("LOG_FORMAT must be %q or %q, got %q", LogFormatText, LogFormatJSON, value)
	}
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
