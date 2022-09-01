package http

import (
	"net/http"

	"github.com/autobrr/filters-go-brr/internal/example"

	"github.com/go-chi/chi/v5"
)

type exampleHandler struct {
	ExampleService *example.Service
}

func newExampleHandler(exampleSvc *example.Service) *exampleHandler {
	return &exampleHandler{
		ExampleService: exampleSvc,
	}
}

func (h exampleHandler) Routes(r chi.Router) {
	r.Get("/", h.hello)
}

func (h exampleHandler) hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(h.ExampleService.Hello()))
}
