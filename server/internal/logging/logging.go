package logging

import (
	"io"
	"log/slog"

	"bowatt/internal/config"
)

func New(cfg config.Config, w io.Writer) *slog.Logger {
	options := &slog.HandlerOptions{Level: cfg.LogLevel}
	if cfg.LogFormat == config.LogFormatJSON {
		return slog.New(slog.NewJSONHandler(w, options))
	}
	return slog.New(slog.NewTextHandler(w, options))
}
