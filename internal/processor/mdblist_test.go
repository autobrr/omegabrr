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

// Unit test for the `mdblist` function with mocked dependencies.
func TestMDBList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return sample JSON response for testing
		fmt.Fprintln(w, `[{"title": "Movie 1"}, {"title": "Movie 2"}]`)
	}))
	defer ts.Close()

	cfg := &domain.ListConfig{
		Name: "test",
		URL:  "https://mdblist.com/lists/linaspurinis/top-watched-movies-of-the-week/json",
	}

	brr := &autobrr.Client{}

	service := Service{}

	err := service.mdblist(context.Background(), cfg, false, brr)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
