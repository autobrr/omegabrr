package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/pkg/autobrr"
	"github.com/rs/zerolog/log"
)

func (s Service) regbrr(ctx context.Context, cfg *domain.ArrConfig, dryRun bool, brr *autobrr.Client) error {
	l := log.With().Str("type", "regbrr").Str("client", cfg.Name).Logger()

	var titles []string

	if strings.HasSuffix(cfg.Host, ".json") {
		l.Debug().Msgf("fetching titles from JSON URL: %s", cfg.Host)

		resp, err := http.Get(cfg.Host)
		if err != nil {
			l.Error().Err(err).Msgf("failed to fetch titles from URL: %s", cfg.Host)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			l.Error().Msgf("failed to fetch titles from URL: %s", cfg.Host)
			return fmt.Errorf("failed to fetch titles from URL: %s", cfg.Host)
		}

		var data []struct {
			Title string `json:"title"`
			Movie struct {
				Title string `json:"title"`
			} `json:"movie"`
			Show struct {
				Title string `json:"title"`
			} `json:"show"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			l.Error().Err(err).Msgf("failed to decode JSON data from URL: %s", cfg.Host)
			return err
		}

		for _, item := range data {
			titles = append(titles, item.Title)
			if item.Movie.Title != "" {
				titles = append(titles, item.Movie.Title)
			}
			if item.Show.Title != "" {
				titles = append(titles, item.Show.Title)
			}
		}

	} else {
		l.Debug().Msgf("fetching titles from URL: %s", cfg.Host)

		resp, err := http.Get(cfg.Host)
		if err != nil {
			l.Error().Err(err).Msgf("failed to fetch titles from URL: %s", cfg.Host)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			l.Error().Msgf("failed to fetch titles from URL: %s", cfg.Host)
			return fmt.Errorf("failed to fetch titles from URL: %s", cfg.Host)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			l.Error().Err(err).Msgf("failed to read response body from URL: %s", cfg.Host)
			return err
		}

		titleLines := strings.Split(string(body), "\n")
		for _, titleLine := range titleLines {
			title := strings.TrimSpace(titleLine)
			if title == "" {
				continue
			}
			titles = append(titles, title)
		}
	}

	//l.Debug().Msgf("gathered titles: %s", strings.ReplaceAll(strings.Join(titles, ", "), ", ", ","))

	for _, filterID := range cfg.Filters {
		l.Debug().Msgf("updating filter: %v", filterID)

		filterTitles := []string{}
		for _, title := range titles {
			filterTitles = append(filterTitles, processTitle(title, cfg.MatchRelease)...)
		}

		joinedTitles := strings.Join(filterTitles, ",")

		l.Trace().Msgf("%s", joinedTitles)

		if len(joinedTitles) == 0 {
			return nil
		}

		f := autobrr.UpdateFilter{Shows: strings.Join(filterTitles, ",")}

		if !dryRun {
			if err := brr.UpdateFilterByID(ctx, filterID, f); err != nil {
				l.Error().Err(err).Msgf("something went wrong updating filter: %v", filterID)
				continue
			}
		}

		l.Debug().Msgf("successfully updated filter: %v", filterID)
	}

	return nil
}
