package processor

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/pkg/autobrr"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/radarr"
)

func (s Service) radarr(ctx context.Context, cfg *domain.ArrConfig, dryRun bool, brr *autobrr.Client) error {
	l := log.With().Str("type", "radarr").Str("client", cfg.Name).Logger()

	l.Debug().Msgf("gathering titles...")

	titles, err := s.processRadarr(ctx, cfg, &l)
	if err != nil {
		return err
	}

	l.Debug().Msgf("got %v filter titles", len(titles))

	joinedTitles := strings.Join(titles, ",")

	l.Trace().Msgf("%v", joinedTitles)

	if len(joinedTitles) == 0 {
		return nil
	}

	for _, filterID := range cfg.Filters {

		l.Debug().Msgf("updating filter: %v", filterID)

		f := autobrr.UpdateFilter{Shows: joinedTitles}

		if cfg.MatchRelease {
			f = autobrr.UpdateFilter{MatchReleases: joinedTitles}
		}

		if !dryRun {
			if err := brr.UpdateFilterByID(ctx, filterID, f); err != nil {
				l.Error().Err(err).Msgf("error updating filter: %v", filterID)
				return errors.Wrapf(err, "error updating filter: %v", filterID)
			}
		}

		l.Debug().Msgf("successfully updated filter: %v", filterID)

	}

	return nil
}

func (s Service) processRadarr(ctx context.Context, cfg *domain.ArrConfig, logger *zerolog.Logger) ([]string, error) {
	c := starr.New(cfg.Apikey, cfg.Host, 60*time.Second)

	if cfg.BasicAuth != nil {
		if cfg.BasicAuth.User != "" {
			c.HTTPUser = cfg.BasicAuth.User
		}
		if cfg.BasicAuth.Pass != "" {
			c.HTTPPass = cfg.BasicAuth.Pass
		}
	}

	r := radarr.New(c)

	var tags []*starr.Tag
	if len(cfg.TagsExclude) > 0 || len(cfg.TagsInclude) > 0 {
		t, err := r.GetTagsContext(ctx)
		if err != nil {
			logger.Debug().Msg("could not get tags")
		}
		tags = t
	}

	movies, err := r.GetMovieContext(ctx, 0)
	if err != nil {
		return nil, err
	}

	logger.Debug().Msgf("found %d movies to process", len(movies))

	titleSet := make(map[string]struct{})
	var processedTitles int

	for _, movie := range movies {
		m := movie

		if !s.shouldProcessItem(m.Monitored, cfg) {
			continue
		}

		if len(cfg.TagsInclude) > 0 {
			if len(m.Tags) == 0 {
				continue
			}
			if !containsTag(tags, m.Tags, cfg.TagsInclude) {
				continue
			}
		}

		if len(cfg.TagsExclude) > 0 {
			if containsTag(tags, m.Tags, cfg.TagsExclude) {
				continue
			}
		}

		processedTitles++

		// Taking the international title and the original title and appending them to the titles array.
		for _, title := range []string{m.Title, m.OriginalTitle} {
			if title != "" {
				for _, t := range processTitle(title, cfg.MatchRelease) {
					titleSet[t] = struct{}{}
				}
			}
		}
	}

	uniqueTitles := make([]string, 0, len(titleSet))
	for title := range titleSet {
		uniqueTitles = append(uniqueTitles, title)
	}

	sort.Strings(uniqueTitles)
	logger.Debug().Msgf("from a total of %d movies we found %d titles and created %d release titles", len(movies), processedTitles, len(uniqueTitles))

	return uniqueTitles, nil
}
