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
	"golang.org/x/sync/errgroup"
	"golift.io/starr"
	"golift.io/starr/radarr"
	"golift.io/starr/sonarr"
)

type Service struct {
	cfg *domain.Config
}

func NewService(cfg *domain.Config) *Service {
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

	start := time.Now()

	var g errgroup.Group

	if s.cfg.Clients.Radarr != nil {
		for _, arrClient := range s.cfg.Clients.Radarr {
			g.Go(func() error {
				if err := s.radarr(ctx, arrClient, dryRun, a); err != nil {
					log.Error().Err(err).Msgf("radarr: %v something went wrong", arrClient.Name)
					return err
				}
				return nil
			})
		}
	}

	if s.cfg.Clients.Sonarr != nil {
		for _, arrClient := range s.cfg.Clients.Sonarr {
			g.Go(func() error {
				if err := s.sonarr(ctx, arrClient, dryRun, a); err != nil {
					log.Error().Err(err).Msgf("sonarr: %v something went wrong", arrClient.Name)
					return err
				}
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		log.Error().Err(err).Msgf("Something went wrong trying to update filters! Total time: %v", time.Since(start))
		return err
	}

	log.Info().Msgf("Successfully updated filters! Total time: %v", time.Since(start))

	return nil
}

func (s Service) radarr(ctx context.Context, cfg *domain.ArrConfig, dryRun bool, brr *autobrr.Client) error {
	l := log.With().Str("type", "sonarr").Str("client", cfg.Name).Logger()

	l.Debug().Msgf("gathering titles...")

	movieTitles, err := s.processRadarr(cfg, l)
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

func (s Service) processRadarr(cfg *domain.ArrConfig, logger zerolog.Logger) ([]string, error) {
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

	logger.Debug().Msgf("found %d movies to process", len(movies))

	var titles []string
	var monitoredTitles int

	for _, m := range movies {
		// only want monitored
		if !m.Monitored {
			continue
		}

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

func (s Service) sonarr(ctx context.Context, cfg *domain.ArrConfig, dryRun bool, brr *autobrr.Client) error {
	l := log.With().Str("type", "sonarr").Str("client", cfg.Name).Logger()

	l.Debug().Msgf("gathering titles...")

	movieTitles, err := s.processSonarr(cfg, l)
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

func (s Service) processSonarr(cfg *domain.ArrConfig, logger zerolog.Logger) ([]string, error) {
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

	logger.Debug().Msgf("found %d shows to process", len(shows))

	var titles []string
	var monitoredTitles int

	for _, m := range shows {
		// only want monitored
		if !m.Monitored {
			continue
		}

		monitoredTitles++

		//titles = append(titles, rls.MustNormalize(m.Title))

		//titles = append(titles, rls.MustClean(m.Title))

		titles = append(titles, processTitle(m.Title)...)

		//for _, title := range m.AlternateTitles {
		//	titles = append(titles, processTitle(title.Title)...)
		//}
	}

	logger.Debug().Msgf("from a total of %d shows we found %d monitored and created %d release titles", len(shows), monitoredTitles, len(titles))

	return titles, nil
}

func (s Service) GetFilters() ([]autobrr.Filter, error) {
	if s.cfg.Clients.Autobrr == nil {
		log.Fatal().Msg("must supply autobrr configuration!")
		return nil, errors.New("must supply autobrr configuration")
	}

	a := autobrr.NewClient(s.cfg.Clients.Autobrr.Host, s.cfg.Clients.Autobrr.Apikey)
	filters, err := a.GetFilters(context.TODO())
	if err != nil {
		return nil, err
	}

	return filters, nil
}
