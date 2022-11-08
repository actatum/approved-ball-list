package abl

import (
	"os"
	"time"

	"cloud.google.com/go/logging"
	"github.com/rs/zerolog"
)

// NewLogger returns a new zerolog.Logger
func NewLogger(service string, lvl zerolog.Level) *zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	logger := zerolog.New(os.Stdout).
		Level(lvl).With().Str("service", service).
		Timestamp().Logger().Hook(&severityHook{})

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
