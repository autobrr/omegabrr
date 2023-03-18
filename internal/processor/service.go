package processor

import (
	"context"
	"fmt"
	"sync"
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

	// Create a slice to store errors
	var processingErrors []string

	// Use a mutex to protect the processingErrors slice
	var mu sync.Mutex

	if s.cfg.Clients.Arr != nil {
		for _, arrClient := range s.cfg.Clients.Arr {
			// https://golang.org/doc/faq#closures_and_goroutines
			arrClient := arrClient

			switch arrClient.Type {
			case domain.ArrTypeRadarr:
				g.Go(func() error {
					if err := s.radarr(ctx, arrClient, dryRun, a); err != nil {
						log.Error().Err(err).Str("type", "radarr").Str("client", arrClient.Name).Msg("error while processing Radarr, continuing with other clients")
						mu.Lock()
						processingErrors = append(processingErrors, fmt.Sprintf("Radarr - %s: %v", arrClient.Name, err))
						mu.Unlock()
					}
					return nil
				})

			case domain.ArrTypeWhisparr:
				g.Go(func() error {
					if err := s.radarr(ctx, arrClient, dryRun, a); err != nil {
						log.Error().Err(err).Str("type", "whisparr").Str("client", arrClient.Name).Msg("error while processing Whisparr, continuing with other clients")
						mu.Lock()
						processingErrors = append(processingErrors, fmt.Sprintf("Whisparr - %s: %v", arrClient.Name, err))
						mu.Unlock()
					}
					return nil
				})

			case domain.ArrTypeSonarr:
				g.Go(func() error {
					if err := s.sonarr(ctx, arrClient, dryRun, a); err != nil {
						log.Error().Err(err).Str("type", "sonarr").Str("client", arrClient.Name).Msg("error while processing Sonarr, continuing with other clients")
						mu.Lock()
						processingErrors = append(processingErrors, fmt.Sprintf("Sonarr - %s: %v", arrClient.Name, err))
						mu.Unlock()
					}
					return nil
				})

			case domain.ArrTypeReadarr:
				g.Go(func() error {
					if err := s.readarr(ctx, arrClient, dryRun, a); err != nil {
						log.Error().Err(err).Str("type", "readarr").Str("client", arrClient.Name).Msg("error while processing Readarr, continuing with other clients")
						mu.Lock()
						processingErrors = append(processingErrors, fmt.Sprintf("Readarr - %s: %v", arrClient.Name, err))
						mu.Unlock()
					}
					return nil
				})

			case domain.ArrTypeLidarr:
				g.Go(func() error {
					if err := s.lidarr(ctx, arrClient, dryRun, a); err != nil {
						log.Error().Err(err).Str("type", "lidarr").Str("client", arrClient.Name).Msg("error while processing Lidarr, continuing with other clients")
						mu.Lock()
						processingErrors = append(processingErrors, fmt.Sprintf("Lidarr - %s: %v", arrClient.Name, err))
						mu.Unlock()
					}
					return nil
				})
			}
		}
	}

	if s.cfg.Clients.Lists != nil {
		for _, listsClient := range s.cfg.Clients.Lists {
			listsClient := listsClient

			switch listsClient.Type {
			case domain.ListTypeTrakt:
				g.Go(func() error {
					if err := s.trakt(ctx, listsClient, dryRun, a); err != nil {
						log.Error().Err(err).Str("type", "trakt").Str("client", listsClient.Name).Msg("error while processing Trakt list, continuing with other lists")
						mu.Lock()
						processingErrors = append(processingErrors, fmt.Sprintf("Trakt - %s: %v", listsClient.Name, err))
						mu.Unlock()
					}
					return nil
				})

			case domain.ListTypeMdblist:
				g.Go(func() error {
					if err := s.mdblist(ctx, listsClient, dryRun, a); err != nil {
						log.Error().Err(err).Str("type", "mdblist").Str("client", listsClient.Name).Msg("error while processing Mdblist, continuing with other lists")
						mu.Lock()
						processingErrors = append(processingErrors, fmt.Sprintf("Mdblist - %s: %v", listsClient.Name, err))
						mu.Unlock()
					}
					return nil
				})

			case domain.ListTypeMetacritic:
				g.Go(func() error {
					if err := s.metacritic(ctx, listsClient, dryRun, a); err != nil {
						log.Error().Err(err).Str("type", "metacritic").Str("client", listsClient.Name).Msg("error while processing Metacritic, continuing with other lists")
						mu.Lock()
						processingErrors = append(processingErrors, fmt.Sprintf("Metacritic - %s: %v", listsClient.Name, err))
						mu.Unlock()
					}
					return nil
				})

			case domain.ListTypePlaintext:
				g.Go(func() error {
					if err := s.plaintext(ctx, listsClient, dryRun, a); err != nil {
						log.Error().Err(err).Str("type", "plaintext").Str("client", listsClient.Name).Msg("error while processing Plaintext list, continuing with other lists")
						mu.Lock()
						processingErrors = append(processingErrors, fmt.Sprintf("Plaintext - %s: %v", listsClient.Name, err))
						mu.Unlock()
					}
					return nil
				})
			}
		}
	}

	if err := g.Wait(); err != nil {
		log.Error().Err(err).Msgf("Something went wrong trying to update filters! Total time: %v", time.Since(start))
		return err
	}

	log.Info().Msgf("Successfully updated filters! Total time: %v", time.Since(start))

	// Print the errors if there are any
	if len(processingErrors) > 0 {
		log.Warn().Msg("Errors encountered during processing:")
		for _, errMsg := range processingErrors {
			log.Warn().Msg(errMsg)
		}
	}

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
