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
	"golift.io/starr/sonarr"
)

func (s Service) sonarr(ctx context.Context, cfg *domain.ArrConfig, dryRun bool, brr *autobrr.Client) error {
	var arrType string
	if cfg.Type == domain.ArrTypeWhisparr {
		arrType = "whisparr"
	} else {
		arrType = "sonarr"
	}

	l := log.With().Str("type", arrType).Str("client", cfg.Name).Logger()

	l.Debug().Msgf("gathering titles...")

	titles, err := s.processSonarr(ctx, cfg, &l)
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

func (s Service) processSonarr(ctx context.Context, cfg *domain.ArrConfig, logger *zerolog.Logger) ([]string, error) {
	c := starr.New(cfg.Apikey, cfg.Host, 60*time.Second)

	if cfg.BasicAuth != nil {
		if cfg.BasicAuth.User != "" {
			c.HTTPUser = cfg.BasicAuth.User
		}
		if cfg.BasicAuth.Pass != "" {
			c.HTTPPass = cfg.BasicAuth.Pass
		}
	}

	r := sonarr.New(c)

	var tags []*starr.Tag
	if len(cfg.TagsExclude) > 0 || len(cfg.TagsInclude) > 0 {
		t, err := r.GetTagsContext(ctx)
		if err != nil {
			logger.Debug().Msg("could not get tags")
		}
		tags = t
	}

	shows, err := r.GetAllSeriesContext(ctx)
	if err != nil {
		return nil, err
	}

	logger.Debug().Msgf("found %d shows to process", len(shows))

	titleSet := make(map[string]struct{})
	var processedTitles int

	for _, show := range shows {
		series := show

		if !s.shouldProcessItem(series.Monitored, cfg) {
			continue
		}

		if len(cfg.TagsInclude) > 0 {
			if len(series.Tags) == 0 {
				continue
			}
			if !containsTag(tags, series.Tags, cfg.TagsInclude) {
				continue
			}
		}

		if len(cfg.TagsExclude) > 0 {
			if containsTag(tags, series.Tags, cfg.TagsExclude) {
				continue
			}
		}

		processedTitles++

		titles := processTitle(series.Title, cfg.MatchRelease)
		for _, title := range titles {
			titleSet[title] = struct{}{}
		}

		if !cfg.ExcludeAlternateTitles {
			for _, title := range series.AlternateTitles {
				altTitles := processTitle(title.Title, cfg.MatchRelease)
				for _, altTitle := range altTitles {
					titleSet[altTitle] = struct{}{}
				}
			}
		}
	}

	uniqueTitles := make([]string, 0, len(titleSet))
	for title := range titleSet {
		uniqueTitles = append(uniqueTitles, title)
	}

	sort.Strings(uniqueTitles)
	logger.Debug().Msgf("from a total of %d shows we found %d titles and created %d release titles", len(shows), processedTitles, len(uniqueTitles))

	return uniqueTitles, nil
}
