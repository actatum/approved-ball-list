package log

import (
	"context"
	"io"
	"log/slog"
	"time"

	"cloud.google.com/go/logging"
	"github.com/lmittmann/tint"
)

// NewLogger returns a new slog logger with optional settings configured.
func NewLogger(out io.Writer, opts ...Option) *slog.Logger {
	options := options{
		level: slog.LevelInfo,
	}

	for _, opt := range opts {
		opt.apply(&options)
	}

	handlerOptions := &slog.HandlerOptions{
		Level: options.level,
	}

	var handler slog.Handler
	handler = slog.NewJSONHandler(out, handlerOptions)
	if options.fmtLog {
		handler = tint.NewHandler(out, &tint.Options{
			Level:      handlerOptions.Level,
			AddSource:  handlerOptions.AddSource,
			TimeFormat: time.Kitchen,
		})
	}

	handler = severityHandler{Handler: handler}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

type severityHandler struct {
	slog.Handler
}

// Handle adds severity field to logs based on log level to help with log filtering in gcp.
func (h severityHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(
		slog.String("severity", levelToSeverity(r.Level).String()),
	)

	return h.Handler.Handle(ctx, r)
}

func levelToSeverity(lvl slog.Level) logging.Severity {
	switch lvl {
	case slog.LevelDebug:
		return logging.Debug

	case slog.LevelInfo:
		return logging.Info

	case slog.LevelWarn:
		return logging.Warning

	case slog.LevelError:
		return logging.Error

	default:
		return logging.Default
	}
}
