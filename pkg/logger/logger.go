package logger

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/traceid"
)

func SetNewLoggerByDefault(logLevel slog.Level) {
	log := slog.New(
		traceid.LogHandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})),
	)
	slog.SetDefault(log)
}

func HTTPLogger(logLevel slog.Level) func(http.Handler) http.Handler {
	return httplog.RequestLogger(slog.Default(), &httplog.Options{
		Level:         logLevel,
		Schema:        httplog.SchemaOTEL.Concise(true),
		RecoverPanics: true,
	})
}
