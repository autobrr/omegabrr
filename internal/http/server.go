package http

import (
	"fmt"
	"net"
	"net/http"

	"github.com/autobrr/filters-go-brr/internal/example"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
)

type Server struct {
	cfg Config

	ExampleService *example.Service
}

type Config struct {
	Host string
	Port int

	ExampleService *example.Service
}

func NewServer(config Config) Server {
	return Server{
		cfg:            config,
		ExampleService: config.ExampleService,
	}
}

func (s Server) Open() error {
	addr := fmt.Sprintf("%v:%v", s.cfg.Host, s.cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "error opening http server")
	}

	server := http.Server{
		Handler: s.Handler(),
	}

	return server.Serve(listener)
}

func (s Server) Handler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Route("/hello", newExampleHandler(s.ExampleService).Routes)

	return r
}
