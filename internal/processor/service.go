package processor

import (
	"context"
	"time"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/pkg/autobrr"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
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
	if s.cfg.Clients.Autobrr == nil {
		log.Fatal().Msg("must supply autobrr configuration!")
		return errors.New("must supply autobrr configuration")
	}

	a := autobrr.NewClient(s.cfg.Clients.Autobrr.Host, s.cfg.Clients.Autobrr.Apikey)
	if s.cfg.Clients.Autobrr.BasicAuth != nil {
		a.SetBasicAuth(s.cfg.Clients.Autobrr.BasicAuth.User, s.cfg.Clients.Autobrr.BasicAuth.Pass)
	}

	log.Debug().Msgf("starting filter processing...")

	start := time.Now()

	g, ctx := errgroup.WithContext(context.Background())

	if s.cfg.Clients.Arr != nil {
		for _, arrClient := range s.cfg.Clients.Arr {
			// https://golang.org/doc/faq#closures_and_goroutines
			arrClient := arrClient

			switch arrClient.Type {
			case domain.ArrTypeRadarr:
				g.Go(func() error {
					return s.radarr(ctx, arrClient, dryRun, a)
				})

			case domain.ArrTypeSonarr:
				g.Go(func() error {
					return s.sonarr(ctx, arrClient, dryRun, a)
				})
			}
		}
	}

	if err := g.Wait(); err != nil {
		log.Error().Err(err).Msgf("Something went wrong trying to update filters! Total time: %v", time.Since(start))
		return err
	}

	log.Info().Msgf("Successfully updated filters! Total time: %v", time.Since(start))

	return nil
}

func (s Service) GetFilters(ctx context.Context) ([]autobrr.Filter, error) {
	if s.cfg.Clients.Autobrr == nil {
		log.Fatal().Msg("must supply autobrr configuration!")
		return nil, errors.New("must supply autobrr configuration")
	}

	a := autobrr.NewClient(s.cfg.Clients.Autobrr.Host, s.cfg.Clients.Autobrr.Apikey)
	if s.cfg.Clients.Autobrr.BasicAuth != nil {
		a.SetBasicAuth(s.cfg.Clients.Autobrr.BasicAuth.User, s.cfg.Clients.Autobrr.BasicAuth.Pass)
	}

	filters, err := a.GetFilters(ctx)
	if err != nil {
		return nil, err
	}

	return filters, nil
}
