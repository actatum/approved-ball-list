package app

import (
	"net/http"
	"time"

	"cloud.google.com/go/logging"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

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

type mwWriter struct {
	msg []byte
}

func (w *mwWriter) Write(p []byte) (n int, err error) {
	w.msg = p
	return len(p), nil
}

func requestLogger(logger *zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			wr := &mwWriter{}
			ww.Tee(wr)

			r = r.WithContext(logger.WithContext(r.Context()))

			defer func(start time.Time) {
				dur := time.Since(start)
				status := ww.Status()

				sublogger := zerolog.Ctx(r.Context()).With().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Str("proto", r.Proto).
					Str("from", r.RemoteAddr).
					Int("status", status).
					Int("size", ww.BytesWritten()).
					Str("duration", dur.String()).
					Logger()

				switch {
				case status >= http.StatusInternalServerError:
					sublogger.Error().Msg("Internal Server Error")
				case status >= http.StatusBadRequest:
					sublogger.Warn().Msg("Client Error")
				case status >= http.StatusMultipleChoices:
					sublogger.Info().Msg("Redirection")
				default:
					sublogger.Info().Msg("Success")
				}
			}(time.Now())

			next.ServeHTTP(ww, r)
		})
	}
}
