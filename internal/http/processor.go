package http

import (
	"github.com/autobrr/omegabrr/internal/processor"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type processorHandler struct {
	ProcessorService *processor.Service
}

func newProcessorHandler(processorSvc *processor.Service) *processorHandler {
	return &processorHandler{
		ProcessorService: processorSvc,
	}
}

func (h processorHandler) Routes(r chi.Router) {
	r.Get("/filters", h.getFilters)
}

func (h processorHandler) getFilters(w http.ResponseWriter, r *http.Request) {
	filters, err := h.ProcessorService.GetFilters()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
	}

	render.JSON(w, r, filters)
}
