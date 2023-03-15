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

func (s Service) trakt(ctx context.Context, cfg *domain.ListConfig, dryRun bool, brr *autobrr.Client) error {
	l := log.With().Str("type", "trakt").Str("client", cfg.Name).Logger()

	// Validate the input URL
	if !strings.HasPrefix(cfg.URL, "https://api.autobrr.com/trakt") {
		return fmt.Errorf("invalid URL provided for Trakt list, URL must start with https://api.autobrr.com/trakt. For supported lists, please refer to the README")
	}

	if cfg.URL == "" {
		errMsg := "no URL provided for Trakt list"
		l.Error().Msg(errMsg)
		return fmt.Errorf(errMsg)
	}

	var titles []string

	green := color.New(color.FgGreen).SprintFunc()
	l.Debug().Msgf("fetching titles from %s", green(cfg.URL))

	resp, err := http.Get(cfg.URL)
	if err != nil {
		l.Error().Err(err).Msgf("failed to fetch titles from URL: %s", cfg.URL)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		l.Error().Msgf("failed to fetch titles from URL: %s", cfg.URL)
		return fmt.Errorf("failed to fetch titles from URL: %s", cfg.URL)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		errMsg := fmt.Sprintf("invalid content type for URL: %s, content type should be application/json", cfg.URL)
		return fmt.Errorf(errMsg)
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
		l.Error().Err(err).Msgf("failed to decode JSON data from URL: %s", cfg.URL)
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

	for _, filterID := range cfg.Filters {
		l.Debug().Msgf("updating filter: %v", filterID)

		filterTitles := []string{}
		for _, title := range titles {
			filterTitles = append(filterTitles, processTitle(title, cfg.MatchRelease)...)
		}

		joinedTitles := strings.Join(filterTitles, ",")

		l.Trace().Msgf("%s", joinedTitles)

		if len(joinedTitles) == 0 {
			if strings.Contains(cfg.URL, "mdblist.com") {
				l.Error().Msgf("Found %s in a Trakt filter", cfg.URL)
				l.Error().Msgf("Please make sure you have set up the URL as \"type: mdblist\" in the config")
			} else {
				l.Error().Msgf("Found no titles in %s", cfg.URL)
				l.Error().Msgf("Are you sure this is a trakt.tv JSON?")
			}
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
