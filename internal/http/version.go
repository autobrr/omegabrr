package http

import (
	"net/http"

	"github.com/autobrr/omegabrr/internal/buildinfo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type versionHandler struct{}

func newVersionHandler() *versionHandler {
	return &versionHandler{}
}

func (h versionHandler) Routes(r chi.Router) {
	r.Get("/", h.handleVersion)
}

type VersionResponse struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

func (h versionHandler) handleVersion(w http.ResponseWriter, r *http.Request) {
	resp := VersionResponse{
		Version: buildinfo.Version,
		Commit:  buildinfo.Commit,
		Date:    buildinfo.Date,
	}
	render.JSON(w, r, resp)
}
