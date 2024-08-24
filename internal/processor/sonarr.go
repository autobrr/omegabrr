package processor

import (
	"context"
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

func (s *Service) sonarr(ctx context.Context, cfg *domain.ArrConfig, dryRun bool, brr *autobrr.Client) error {
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

func (s *Service) processSonarr(ctx context.Context, cfg *domain.ArrConfig, logger *zerolog.Logger) ([]string, error) {
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

	var titles []string
	var monitoredTitles int

	for _, show := range shows {
		s := show

		// only want monitored
		if !s.Monitored {
			continue
		}

		if len(cfg.TagsInclude) > 0 {
			if len(s.Tags) == 0 {
				continue
			}
			if !containsTag(tags, s.Tags, cfg.TagsInclude) {
				continue
			}
		}

		if len(cfg.TagsExclude) > 0 {
			if containsTag(tags, s.Tags, cfg.TagsExclude) {
				continue
			}
		}

		// increment monitored titles
		monitoredTitles++

		//titles = append(titles, rls.MustNormalize(s.Title))
		//titles = append(titles, rls.MustClean(s.Title))

		titles = append(titles, processTitle(s.Title, cfg.MatchRelease)...)

		if !cfg.ExcludeAlternateTitles {
			for _, title := range s.AlternateTitles {
				titles = append(titles, processTitle(title.Title, cfg.MatchRelease)...)
			}
		}
	}

	logger.Debug().Msgf("from a total of %d shows we found %d monitored and created %d release titles", len(shows), monitoredTitles, len(titles))

	return titles, nil
}
