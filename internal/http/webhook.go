package http

import (
	"net/http"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/internal/processor"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type webhookHandler struct {
	cfg              *domain.Config
	processorService *processor.Service
}

func newWebhookHandler(cfg *domain.Config, processorSvc *processor.Service) *webhookHandler {
	return &webhookHandler{
		cfg:              cfg,
		processorService: processorSvc,
	}
}

func (h *webhookHandler) Routes(r chi.Router) {
	r.Get("/trigger", h.run)
	r.Post("/trigger", h.run)
	r.Get("/trigger/arr", h.arr)
	r.Get("/trigger/lists", h.lists)
	r.Post("/trigger/arr", h.arr)
	r.Post("/trigger/lists", h.lists)
}

func (h *webhookHandler) run(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	errArrs := h.processorService.ProcessArrs(ctx, false)
	errLists := h.processorService.ProcessLists(ctx, false)

	if errArrs != nil || errLists != nil {
		render.NoContent(w, r)
		return
	}

	render.Status(r, http.StatusOK)
}

func (h *webhookHandler) arr(w http.ResponseWriter, r *http.Request) {
	if err := h.processorService.ProcessArrs(r.Context(), false); err != nil {
		render.NoContent(w, r)
		return
	}

	render.Status(r, http.StatusOK)
}

func (h *webhookHandler) lists(w http.ResponseWriter, r *http.Request) {
	if err := h.processorService.ProcessLists(r.Context(), false); err != nil {
		render.NoContent(w, r)
		return
	}

	render.Status(r, http.StatusOK)
}
