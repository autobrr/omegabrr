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
	"golift.io/starr/readarr"
)

func (s *Service) readarr(ctx context.Context, cfg *domain.ArrConfig, dryRun bool, brr *autobrr.Client) error {
	l := log.With().Str("type", "readarr").Str("client", cfg.Name).Logger()

	l.Debug().Msgf("gathering titles...")

	titles, err := s.processReadarr(ctx, cfg, &l)
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

		if !dryRun {
			f := autobrr.UpdateFilter{MatchReleases: joinedTitles}

			if !dryRun {
				if err := brr.UpdateFilterByID(ctx, filterID, f); err != nil {
					l.Error().Err(err).Msgf("error updating filter: %v", filterID)
					return errors.Wrapf(err, "error updating filter: %v", filterID)
				}
			}
		}

		l.Debug().Msgf("successfully updated filter: %v", filterID)
	}

	return nil
}

func (s *Service) processReadarr(ctx context.Context, cfg *domain.ArrConfig, logger *zerolog.Logger) ([]string, error) {
	c := starr.New(cfg.Apikey, cfg.Host, 60*time.Second)

	if cfg.BasicAuth != nil {
		if cfg.BasicAuth.User != "" {
			c.HTTPUser = cfg.BasicAuth.User
		}
		if cfg.BasicAuth.Pass != "" {
			c.HTTPPass = cfg.BasicAuth.Pass
		}
	}

	r := readarr.New(c)

	// I did not find support for tags here.
	//
	//	var tags []*starr.Tag
	//	if len(cfg.TagsExclude) > 0 || len(cfg.TagsInclude) > 0 {
	//		t, err := r.GetTagsContext(ctx)
	//		if err != nil {
	//			logger.Debug().Msg("could not get tags")
	//		}
	//		tags = t
	//	}

	ebooks, err := r.GetBookContext(ctx, "")
	if err != nil {
		return nil, err
	}

	logger.Debug().Msgf("found %d ebooks to process", len(ebooks))

	var titles []string
	var monitoredTitles int

	for _, ebook := range ebooks {
		m := ebook

		// only want monitored
		if !m.Monitored {
			continue
		}

		// I did not find support for tags here
		//
		//		if len(cfg.TagsInclude) > 0 {
		//			if len(m.Tags) == 0 {
		//				continue
		//			}
		//			if !containsTag(tags, m.Tags, cfg.TagsInclude) {
		//				continue
		//			}
		//		}
		//
		//		if len(cfg.TagsExclude) > 0 {
		//			if containsTag(tags, m.Tags, cfg.TagsExclude) {
		//				continue
		//			}
		//		}

		// increment monitored titles
		monitoredTitles++

		//titles = append(titles, rls.MustNormalize(m.Title))
		//titles = append(titles, rls.MustNormalize(m.OriginalTitle))
		//titles = append(titles, rls.MustClean(m.Title))

		titles = append(titles, processTitle(m.Title, cfg.MatchRelease)...)

		//	titles = append(titles, processTitle(m.OriginalTitle)...)

		//for _, title := range m.AlternateTitles {
		//	titles = append(titles, processTitle(title.Title)...)
		//}
	}

	logger.Debug().Msgf("from a total of %d ebooks we found %d monitored and created %d release titles", len(ebooks), monitoredTitles, len(titles))

	return titles, nil
}
