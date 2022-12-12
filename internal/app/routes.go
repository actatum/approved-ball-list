package app

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/actatum/approved-ball-list/internal/abl"

	"github.com/actatum/errs/httperr"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type handler struct {
	svc    abl.Service
	logger *zerolog.Logger
	cfg    config
}

type envelope map[string]any

func (h *handler) routes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		requestLogger(h.logger),
		middleware.StripSlashes,
		middleware.Recoverer,
	)

	r.Get("/v1/health", h.handleHealth())
	r.Get("/v1/cron", h.handleCronJob())

	return r
}

func (h *handler) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := envelope{
			"status": "available",
			"system_info": map[string]string{
				"environment": h.cfg.Env,
				"version":     "v1",
			},
		}

		err := h.renderJSON(w, http.StatusOK, response, nil)
		if err != nil {
			zerolog.Ctx(r.Context()).Error().Err(err).Send()
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func (h *handler) handleCronJob() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.svc.RefreshBalls(r.Context())
		if err != nil {
			zerolog.Ctx(r.Context()).Error().Err(err).Send()
			// sentry it?
			httperr.RenderError(err, w)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *handler) renderJSON(w http.ResponseWriter, status int, response envelope, headers http.Header) error {
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
