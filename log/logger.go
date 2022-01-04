package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger returns a new zap logger with the configured logging level
func NewLogger() (*zap.Logger, error) {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.DisableCaller = true
	loggerConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	logger, err := loggerConfig.Build(zap.AddStacktrace(zap.FatalLevel))
	if err != nil {
		return nil, err
	}

	return logger, nil
}
