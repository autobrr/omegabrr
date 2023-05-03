package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/pkg/autobrr"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
)

func (s Service) mdblist(ctx context.Context, cfg *domain.ListConfig, dryRun bool, brr *autobrr.Client) error {
	l := log.With().Str("type", "mdblist").Str("client", cfg.Name).Logger()

	if cfg.URL == "" {
		errMsg := "no URL provided for Mdblist"
		l.Error().Msg(errMsg)
		return fmt.Errorf(errMsg)
	}

	var titles []string

	green := color.New(color.FgGreen).SprintFunc()
	l.Debug().Msgf("fetching titles from %s", green(cfg.URL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL, nil)
	if err != nil {
		l.Error().Err(err).Msg("could not make new request")
		return err
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		l.Error().Err(err).Msgf("failed to fetch titles from URL: %s", cfg.URL)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		l.Error().Msgf("failed to fetch titles from URL: %s", cfg.URL)
		return fmt.Errorf("failed to fetch titles from URL: %s", cfg.URL)
	}

	var data []struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		l.Error().Err(err).Msgf("failed to decode JSON data from URL: %s", cfg.URL)
		return err
	}

	for _, item := range data {
		titles = append(titles, item.Title)
	}

	for _, filterID := range cfg.Filters {
		l.Debug().Msgf("updating filter: %v", filterID)

		filterTitles := []string{}
		for _, title := range titles {
			filterTitles = append(filterTitles, processTitle(title, cfg.MatchRelease)...)
		}

		joinedTitles := strings.Join(filterTitles, ",")

		l.Trace().Msgf("%s", joinedTitles)

		if len(joinedTitles) == 0 {
			l.Debug().Msgf("no titles found for filter: %v", filterID)
			return nil
		}

		f := autobrr.UpdateFilter{Shows: strings.Join(filterTitles, ",")}

		if !dryRun {
			if err := brr.UpdateFilterByID(ctx, filterID, f); err != nil {
				errMsg := fmt.Sprintf("error updating filter: %v, %v", filterID, err)
				l.Error().Msg(errMsg)
				return fmt.Errorf("%s", errMsg)
			}
		}

		l.Debug().Msgf("successfully updated filter: %v", filterID)
	}

	return nil
}
