package processor

import (
	"context"
	"strings"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/pkg/autobrr"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golift.io/starr"
	"golift.io/starr/radarr"
	"golift.io/starr/sonarr"
)

type Service struct {
	cfg domain.Config
}

func NewService(cfg domain.Config) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s Service) Process(dryRun bool) error {
	ctx := context.TODO()

	if s.cfg.Clients.Autobrr == nil {
		log.Fatal().Msg("must supply autobrr configuration!")
		return errors.New("must supply autobrr configuration")
	}

	a := autobrr.NewClient(s.cfg.Clients.Autobrr.Host, s.cfg.Clients.Autobrr.Apikey)

	log.Debug().Msgf("starting filter processing...")

	if s.cfg.Clients.Radarr != nil {
		log.Debug().Msgf("radarr: gathering titles...")

		movieTitles, err := s.processRadarr(s.cfg.Clients.Radarr)
		if err != nil {
			return err
		}

		log.Debug().Msgf("radarr: got %v titles", len(movieTitles))

		joinedTitles := strings.Join(movieTitles, ",")

		log.Trace().Msgf("%v", joinedTitles)

		for _, filterID := range s.cfg.Clients.Radarr.Filters {

			log.Debug().Msgf("radarr: updating filter: %v", filterID)

			if !dryRun {
				f := autobrr.UpdateFilter{MatchReleases: joinedTitles}

				if err := a.UpdateFilterByID(ctx, filterID, f); err != nil {
					log.Error().Err(err).Msgf("radarr: something went wrong updating movie filter: %v", filterID)
					continue
				}
			}

			log.Debug().Msgf("radarr: successfully updated filter: %v", filterID)
		}
	}

	if s.cfg.Clients.Radarr != nil {
		log.Debug().Msgf("sonarr: gathering titles...")

		tvTitles, err := s.processSonarr(s.cfg.Clients.Sonarr)
		if err != nil {
			return err
		}

		log.Debug().Msgf("sonarr: got %v titles", len(tvTitles))

		joinedTitles := strings.Join(tvTitles, ",")

		log.Trace().Msgf("%v", joinedTitles)

		for _, filterID := range s.cfg.Clients.Sonarr.Filters {

			log.Debug().Msgf("sonarr: updating filter: %v", filterID)

			if !dryRun {
				f := autobrr.UpdateFilter{MatchReleases: joinedTitles}

				if err := a.UpdateFilterByID(ctx, filterID, f); err != nil {
					log.Error().Err(err).Msgf("sonarr: something went wrong updating tv filter: %v", filterID)
					continue
				}
			}

			log.Debug().Msgf("sonarr: successfully updated filter: %v", filterID)
		}
	}

	log.Info().Msgf("Successfully updated filters!")

	return nil
}

func (s Service) processRadarr(cfg *domain.ArrConfig) ([]string, error) {
	c := starr.New(cfg.Apikey, cfg.Host, 0)

	if cfg.BasicAuth != nil {
		if cfg.BasicAuth.User != "" {
			c.HTTPUser = cfg.BasicAuth.User
		}
		if cfg.BasicAuth.Pass != "" {
			c.HTTPPass = cfg.BasicAuth.Pass
		}
	}

	r := radarr.New(c)

	movies, err := r.GetMovie(0)
	if err != nil {
		return nil, err
	}

	var titles []string

	for _, m := range movies {
		// only want monitored
		if !m.Monitored {
			continue
		}

		//titles = append(titles, rls.MustNormalize(m.Title))
		//titles = append(titles, rls.MustNormalize(m.OriginalTitle))
		//titles = append(titles, rls.MustClean(m.Title))

		titles = append(titles, processTitle(m.Title)...)
		titles = append(titles, processTitle(m.OriginalTitle)...)

		//for _, title := range m.AlternateTitles {
		//	titles = append(titles, processTitle(title.Title)...)
		//}
	}

	return titles, nil
}

func (s Service) processSonarr(cfg *domain.ArrConfig) ([]string, error) {
	c := starr.New(cfg.Apikey, cfg.Host, 0)

	if cfg.BasicAuth != nil {
		if cfg.BasicAuth.User != "" {
			c.HTTPUser = cfg.BasicAuth.User
		}
		if cfg.BasicAuth.Pass != "" {
			c.HTTPPass = cfg.BasicAuth.Pass
		}
	}

	r := sonarr.New(c)

	shows, err := r.GetAllSeries()
	if err != nil {
		return nil, err
	}

	var titles []string

	for _, m := range shows {
		// only want monitored
		if !m.Monitored {
			continue
		}

		//titles = append(titles, rls.MustNormalize(m.Title))

		//titles = append(titles, rls.MustClean(m.Title))

		titles = append(titles, processTitle(m.Title)...)

		//for _, title := range m.AlternateTitles {
		//	titles = append(titles, processTitle(title.Title)...)
		//}
	}

	return titles, nil
}
