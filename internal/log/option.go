package log

import "log/slog"

type Option interface {
	apply(*options)
}

type options struct {
	fmtLog bool
	level  slog.Level
}

type fmtLogOption struct{}

func (o fmtLogOption) apply(opts *options) {
	opts.fmtLog = true
}

func WithFmtLog() Option {
	return fmtLogOption{}
}

type levelOption slog.Level

func (o levelOption) apply(opts *options) {
	opts.level = slog.Level(o)
}

func WithLevel(lvl slog.Level) Option {
	return levelOption(lvl)
}
