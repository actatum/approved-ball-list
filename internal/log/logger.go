package log

import (
	"io"
	"log/slog"
	"time"

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

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
