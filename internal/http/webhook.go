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

func (h webhookHandler) Routes(r chi.Router) {
	r.Get("/trigger", h.run)
}

func (h webhookHandler) run(w http.ResponseWriter, r *http.Request) {
	if err := h.processorService.Process(false); err != nil {
		render.NoContent(w, r)
	}
	render.Status(r, http.StatusOK)
}
