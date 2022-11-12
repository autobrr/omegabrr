package processor

import (
	"context"
	"strings"
	"time"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/pkg/autobrr"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/radarr"
)

func (s Service) radarr(ctx context.Context, cfg *domain.ArrConfig, dryRun bool, brr *autobrr.Client) error {
	l := log.With().Str("type", "radarr").Str("client", cfg.Name).Logger()

	l.Debug().Msgf("gathering titles...")

	movieTitles, err := s.processRadarr(ctx, cfg, &l)
	if err != nil {
		return err
	}

	l.Debug().Msgf("got %v filter titles", len(movieTitles))

	joinedTitles := strings.Join(movieTitles, ",")

	l.Trace().Msgf("%v", joinedTitles)

	if len(joinedTitles) == 0 {
		return nil
	}

	for _, filterID := range cfg.Filters {

		l.Debug().Msgf("updating filter: %v", filterID)

		if !dryRun {
			f := autobrr.UpdateFilter{MatchReleases: joinedTitles}

			if err := brr.UpdateFilterByID(ctx, filterID, f); err != nil {
				l.Error().Err(err).Msgf("something went wrong updating movie filter: %v", filterID)
				continue
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

	var titles []string
	var monitoredTitles int

	for _, movie := range movies {
		m := movie

		// only want monitored
		if !m.Monitored {
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

		// increment monitored titles
		monitoredTitles++

		//titles = append(titles, rls.MustNormalize(m.Title))
		//titles = append(titles, rls.MustNormalize(m.OriginalTitle))
		//titles = append(titles, rls.MustClean(m.Title))

		t := strings.ToLower(m.Title)
		ot := strings.ToLower(m.OriginalTitle)

		if t == ot {
			titles = append(titles, processTitle(m.Title)...)
			continue
		}

		titles = append(titles, processTitle(m.OriginalTitle)...)

		//for _, title := range m.AlternateTitles {
		//	titles = append(titles, processTitle(title.Title)...)
		//}
	}

	logger.Debug().Msgf("from a total of %d movies we found %d monitored and created %d release titles", len(movies), monitoredTitles, len(titles))

	return titles, nil
}
