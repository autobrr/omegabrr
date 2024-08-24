package http

import (
	"fmt"
	"net"
	"net/http"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/internal/processor"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
)

type Server struct {
	cfg *domain.Config

	processorService *processor.Service
}

func NewServer(config *domain.Config, processorService *processor.Service) *Server {
	return &Server{
		cfg:              config,
		processorService: processorService,
	}
}

func (s *Server) Open() error {
	addr := fmt.Sprintf("%v:%v", s.cfg.Server.Host, s.cfg.Server.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "error opening http server")
	}

	server := http.Server{
		Handler: s.Handler(),
	}

	return server.Serve(listener)
}

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Route("/api/healthz", newHealthHandler().Routes)

	r.Group(func(r chi.Router) {
		r.Use(s.isAuthenticated)

		r.Route("/api", func(r chi.Router) {
			r.Route("/", newProcessorHandler(s.processorService).Routes)
			r.Route("/webhook", newWebhookHandler(s.cfg, s.processorService).Routes)
		})
	})

	return r
}
