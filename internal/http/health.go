package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type healthHandler struct{}

func newHealthHandler() *healthHandler {
	return &healthHandler{}
}

func (h *healthHandler) Routes(r chi.Router) {
	r.Get("/liveness", h.handleLiveness)
	r.Get("/readiness", h.handleReadiness)
}

func (h *healthHandler) handleLiveness(w http.ResponseWriter, r *http.Request) {
	writeHealthy(w, r)
}

func (h *healthHandler) handleReadiness(w http.ResponseWriter, r *http.Request) {
	writeHealthy(w, r)
}

func writeHealthy(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "OK")
}

func writeUnhealthy(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	render.PlainText(w, r, "Unhealthy")
}
