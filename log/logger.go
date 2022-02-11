package log

import (
	"io"
	"os"
	"time"

	"cloud.google.com/go/logging"
	"github.com/rs/zerolog"
)

// NewLogger returns a new zerolog.Logger
func NewLogger(service string, lvl zerolog.Level, opts ...Option) *zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	options := loggerOptions{writer: os.Stdout}

	for _, o := range opts {
		o.apply(&options)
	}

	logger := zerolog.New(options.writer).Level(lvl).With().Str("service", service).Timestamp().Logger().Hook(&severityHook{})

	return &logger
}

var logLevelMappings = map[zerolog.Level]logging.Severity{
	zerolog.NoLevel:    logging.Default,
	zerolog.TraceLevel: logging.Default,
	zerolog.DebugLevel: logging.Debug,
	zerolog.InfoLevel:  logging.Info,
	zerolog.WarnLevel:  logging.Warning,
	zerolog.ErrorLevel: logging.Error,
	zerolog.FatalLevel: logging.Critical,
	zerolog.PanicLevel: logging.Critical,
}

type severityHook struct{}

func (h severityHook) Run(e *zerolog.Event, level zerolog.Level, _ string) {
	e.Str("severity", logLevelMappings[level].String())
}

type loggerOptions struct {
	writer io.Writer
}

// Option represents functional options to configure the Consumer and Producer
type Option interface {
	apply(options *loggerOptions)
}

type writerOption struct {
	w io.Writer
}

func (w writerOption) apply(opts *loggerOptions) {
	opts.writer = w.w
}

// WithWriter configures the number of goroutines to distribute consumed messages to
func WithWriter(w io.Writer) Option {
	return writerOption{w: w}
}
