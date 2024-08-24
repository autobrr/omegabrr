package http

import (
	"net/http"

	"github.com/autobrr/omegabrr/internal/processor"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type processorHandler struct {
	processorService *processor.Service
}

func newProcessorHandler(processorSvc *processor.Service) *processorHandler {
	return &processorHandler{
		processorService: processorSvc,
	}
}

func (h *processorHandler) Routes(r chi.Router) {
	r.Get("/filters", h.getFilters)
}

func (h *processorHandler) getFilters(w http.ResponseWriter, r *http.Request) {
	filters, err := h.processorService.GetFilters(r.Context())
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, filters)
}
