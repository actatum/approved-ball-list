package app

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/actatum/errs/httperr"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type envelope map[string]any

func (a *Application) routes(logger *zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		requestLogger(logger),
		middleware.StripSlashes,
		middleware.Recoverer,
	)

	r.Get("/v1/health", a.handleHealth())
	r.Get("/v1/cron", a.handleCronJob())

	return r
}

func (a *Application) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := envelope{
			"status": "available",
			"system_info": map[string]string{
				"environment": a.config.Env,
				"version":     "v1",
			},
		}

		err := a.renderJSON(w, http.StatusOK, response, nil)
		if err != nil {
			zerolog.Ctx(r.Context()).Error().Err(err).Send()
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func (a *Application) handleCronJob() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := a.service.RefreshBalls(r.Context())
		if err != nil {
			zerolog.Ctx(r.Context()).Error().Err(err).Send()
			// sentry it?
			httperr.RenderError(err, w)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (a *Application) renderJSON(w http.ResponseWriter, status int, response envelope, headers http.Header) error {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(response); err != nil {
		return err
	}

	for k, v := range headers {
		w.Header()[k] = v
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, err := w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
