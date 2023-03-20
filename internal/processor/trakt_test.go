package processor

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/pkg/autobrr"
)

// Tests that an error is returned when an invalid URL is passed to the `trakt` function
func TestTrakt_InvalidURL(t *testing.T) {

	s := Service{}

	cfg := &domain.ListConfig{Name: "test list", URL: "http://example.com"}

	err := s.trakt(context.Background(), cfg, false, nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// Unit test for the `trakt` function with mocked dependencies.
func TestTraktList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return sample JSON response for testing
		fmt.Fprintln(w, `[{"title": "Movie 1", "movie": {"title": "Movie 1 Title"}}, {"title": "Movie 2", "show": {"title": "Show 1 Title"}}]`)
	}))
	defer ts.Close()

	cfg := &domain.ListConfig{
		Name: "test",
		URL:  "https://sudoer.dev/ant.json",
	}

	brr := &autobrr.Client{}

	service := Service{}

	err := service.trakt(context.Background(), cfg, false, brr)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
