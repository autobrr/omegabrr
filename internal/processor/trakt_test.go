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

// Unit test for the `trakt` function with mocked dependencies.
func TestTraktList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return sample JSON response for testing
		fmt.Fprintln(w, `[{"title": "Movie 1", "movie": {"title": "Movie 1 Title"}}, {"title": "Movie 2", "show": {"title": "Show 1 Title"}}]`)
	}))
	defer ts.Close()

	cfg := &domain.ListConfig{
		Name: "test",
		URL:  "https://api.autobrr.com/lists/trakt/anticipated-tv",
	}

	brr := &autobrr.Client{}

	service := Service{}

	err := service.trakt(context.Background(), cfg, false, brr)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
