package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/pkg/autobrr"
	"github.com/rs/zerolog/log"
)

func (s Service) steam(ctx context.Context, cfg *domain.ListConfig, dryRun bool, brr *autobrr.Client) error {
	l := log.With().Str("type", "steam").Str("client", cfg.Name).Logger()

	if cfg.URL == "" {
		errMsg := "no URL provided for Steam wishlist"
		l.Error().Msg(errMsg)
		return fmt.Errorf(errMsg)
	}

	l.Debug().Msgf("fetching titles from %s", cfg.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL, nil)
	if err != nil {
		l.Error().Err(err).Msg("could not create new request")
		return err
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	setUserAgent(req)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		l.Error().Err(err).Msg("failed to fetch titles")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		l.Error().Msg("failed to fetch titles, non-OK HTTP status received")
		return fmt.Errorf("failed to fetch titles, non-OK HTTP status received")
	}

	var data map[string]struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		l.Error().Err(err).Msg("failed to decode JSON data")
		return err
	}

	var titles []string
	for _, item := range data {
		titles = append(titles, item.Name)
	}

	for _, filterID := range cfg.Filters {
		filterTitles := []string{}
		for _, title := range titles {
			filterTitles = append(filterTitles, processTitle(title, cfg.MatchRelease)...)
		}

		joinedTitles := strings.Join(filterTitles, ",")

		l.Trace().Msgf("%s", joinedTitles)

		if len(joinedTitles) == 0 {
			l.Debug().Msgf("no titles found for filter: %v", filterID)
			continue
		}

		f := autobrr.UpdateFilter{MatchReleases: joinedTitles}

		if !dryRun {
			if err := brr.UpdateFilterByID(ctx, filterID, f); err != nil {
				l.Error().Err(err).Msgf("error updating filter: %v", filterID)
				return err
			}
		}

		l.Debug().Msgf("successfully updated filter: %v", filterID)
	}

	return nil
}
