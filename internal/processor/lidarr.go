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
	"golift.io/starr/lidarr"
)

func (s Service) lidarr(ctx context.Context, cfg *domain.ArrConfig, dryRun bool, brr *autobrr.Client) error {
	l := log.With().Str("type", "sonarr").Str("client", cfg.Name).Logger()

	l.Debug().Msgf("gathering titles...")

	titles, err := s.processLidarr(ctx, cfg, &l)
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

		f := autobrr.UpdateFilter{Albums: joinedTitles}
		s := autobrr.UpdateFilterSpecial{Albums: joinedTitles}

		if cfg.MatchRelease {
			f = autobrr.UpdateFilter{MatchReleases: joinedTitles}

			if !dryRun {
				if err := brr.UpdateFilterByID(ctx, filterID, f); err != nil {
					l.Error().Err(err).Msgf("something went wrong updating tv filter: %v", filterID)
					continue
				}
			}

		} else if !cfg.KeepReleaseData {
			f = autobrr.UpdateFilter{Albums: joinedTitles}

			if !dryRun {
				if err := brr.UpdateFilterByID(ctx, filterID, f); err != nil {
					l.Error().Err(err).Msgf("something went wrong updating tv filter: %v", filterID)
					continue
				}
			}
		} else if cfg.KeepReleaseData && !cfg.MatchRelease {
			s = autobrr.UpdateFilterSpecial{Albums: joinedTitles}

			if !dryRun {
				if err := brr.UpdateFilterSpecial(ctx, filterID, s); err != nil {
					l.Error().Err(err).Msgf("something went wrong updating tv filter: %v", filterID)
					continue
				}
			}
		}

		l.Debug().Msgf("successfully updated filter: %v", filterID)

	}

	return nil
}

func (s Service) processLidarr(ctx context.Context, cfg *domain.ArrConfig, logger *zerolog.Logger) ([]string, error) {
	c := starr.New(cfg.Apikey, cfg.Host, 60*time.Second)

	if cfg.BasicAuth != nil {
		if cfg.BasicAuth.User != "" {
			c.HTTPUser = cfg.BasicAuth.User
		}
		if cfg.BasicAuth.Pass != "" {
			c.HTTPPass = cfg.BasicAuth.Pass
		}
	}

	r := lidarr.New(c)

	//		TAGS NOT SUPPORTED FOR ALBUMS APPARENTLY
	//
	//	var tags []*starr.Tag
	//	if len(cfg.TagsExclude) > 0 || len(cfg.TagsInclude) > 0 {
	//		t, err := r.GetTagsContext(ctx)
	//		if err != nil {
	//			logger.Debug().Msg("could not get tags")
	//		}
	//		tags = t
	//	}

	albums, err := r.GetAlbumContext(ctx, "")
	if err != nil {
		return nil, err
	}

	logger.Debug().Msgf("found %d releases to process", len(albums))

	var titles []string
	var monitoredTitles int

	for _, album := range albums {
		m := album

		// only want monitored
		if !m.Monitored {
			continue
		}

		//		TAGS NOT SUPPORTED FOR ALBUMS APPARENTLY
		//
		//		if len(cfg.TagsInclude) > 0 {
		//			if len(s.Tags) == 0 {
		//				continue
		//			}
		//			if !containsTag(tags, s.Tags, cfg.TagsInclude) {
		//				continue
		//			}
		//		}
		//
		//		if len(cfg.TagsExclude) > 0 {
		//			if containsTag(tags, s.Tags, cfg.TagsExclude) {
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

	logger.Debug().Msgf("from a total of %d releases we found %d monitored and created %d release titles", len(albums), monitoredTitles, len(titles))

	return titles, nil
}
