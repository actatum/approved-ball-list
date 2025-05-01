package balls

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func NewHTTPHandler(logger *slog.Logger, svc Service, env string) http.Handler {
	r := chi.NewRouter()

	r.Use(
		middleware.StripSlashes,
		requestLogger(logger),
		middleware.Recoverer,
	)

	r.Get("/v1/health", handleHealth(env))
	r.Get("/v1/cron", handleCron(logger, svc))

	return r
}

func handleHealth(env string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, map[string]any{
			"status": "available",
			"system_info": map[string]string{
				"environment": env,
				"version":     "v1",
			},
		})
	}
}

func handleCron(logger *slog.Logger, svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := svc.CheckForNewlyApprovedBalls(r.Context())
		if err != nil {
			logger.ErrorContext(r.Context(), "error checking for newly approved balls", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]any{
				"error": map[string]any{
					"message": "internal server error",
				},
			})
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func requestLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func(start time.Time) {
				attrs := []slog.Attr{
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("proto", r.Proto),
					slog.String("from", r.RemoteAddr),
					slog.Int("status", ww.Status()),
					slog.Int("size", ww.BytesWritten()),
					slog.String("duration", time.Since(start).String()),
				}

				lvl := slog.LevelInfo
				switch {
				case ww.Status() >= http.StatusInternalServerError:
					lvl = slog.LevelError

				case ww.Status() >= http.StatusBadRequest:
					lvl = slog.LevelWarn
				}

				logger.LogAttrs(r.Context(), lvl, "", attrs...)
			}(time.Now())

			next.ServeHTTP(ww, r)
		})
	}
}
