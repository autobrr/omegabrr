package processor

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/pkg/autobrr"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	cfg *domain.Config

	httpClient *http.Client
}

func NewService(cfg *domain.Config) *Service {
	return &Service{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s Service) ProcessArrs(ctx context.Context, dryRun bool, a *autobrr.Client) []string {
	var processingErrors []string

	if s.cfg.Clients.Arr != nil {
		for _, arrClient := range s.cfg.Clients.Arr {
			arrClient := arrClient

			switch arrClient.Type {
			case domain.ArrTypeRadarr:
				if err := s.radarr(ctx, arrClient, dryRun, a); err != nil {
					log.Error().Err(err).Str("type", "radarr").Str("client", arrClient.Name).Msg("error while processing Radarr, continuing with other clients")
					processingErrors = append(processingErrors, fmt.Sprintf("Radarr - %s: %v", arrClient.Name, err))
				}

			case domain.ArrTypeWhisparr:
				if err := s.radarr(ctx, arrClient, dryRun, a); err != nil {
					log.Error().Err(err).Str("type", "whisparr").Str("client", arrClient.Name).Msg("error while processing Whisparr, continuing with other clients")
					processingErrors = append(processingErrors, fmt.Sprintf("Whisparr - %s: %v", arrClient.Name, err))
				}

			case domain.ArrTypeSonarr:
				if err := s.sonarr(ctx, arrClient, dryRun, a); err != nil {
					log.Error().Err(err).Str("type", "sonarr").Str("client", arrClient.Name).Msg("error while processing Sonarr, continuing with other clients")
					processingErrors = append(processingErrors, fmt.Sprintf("Sonarr - %s: %v", arrClient.Name, err))
				}

			case domain.ArrTypeReadarr:
				if err := s.readarr(ctx, arrClient, dryRun, a); err != nil {
					log.Error().Err(err).Str("type", "readarr").Str("client", arrClient.Name).Msg("error while processing Readarr, continuing with other clients")
					processingErrors = append(processingErrors, fmt.Sprintf("Readarr - %s: %v", arrClient.Name, err))
				}

			case domain.ArrTypeLidarr:
				if err := s.lidarr(ctx, arrClient, dryRun, a); err != nil {
					log.Error().Err(err).Str("type", "lidarr").Str("client", arrClient.Name).Msg("error while processing Lidarr, continuing with other clients")
					processingErrors = append(processingErrors, fmt.Sprintf("Lidarr - %s: %v", arrClient.Name, err))
				}
			}
		}
	}

	return processingErrors
}

func (s Service) ProcessLists(ctx context.Context, dryRun bool, a *autobrr.Client) []string {
	var processingErrors []string

	if s.cfg.Lists != nil {
		for _, listsClient := range s.cfg.Lists {
			listsClient := listsClient

			switch listsClient.Type {
			case domain.ListTypeTrakt:
				if err := s.trakt(ctx, listsClient, dryRun, a); err != nil {
					log.Error().Err(err).Str("type", "trakt").Str("client", listsClient.Name).Msg("error while processing Trakt list, continuing with other lists")
					processingErrors = append(processingErrors, fmt.Sprintf("Trakt - %s: %v", listsClient.Name, err))
				}

			case domain.ListTypeMdblist:
				if err := s.mdblist(ctx, listsClient, dryRun, a); err != nil {
					log.Error().Err(err).Str("type", "mdblist").Str("client", listsClient.Name).Msg("error while processing Mdblist, continuing with other lists")
					processingErrors = append(processingErrors, fmt.Sprintf("Mdblist - %s: %v", listsClient.Name, err))
				}

			case domain.ListTypeMetacritic:
				if err := s.metacritic(ctx, listsClient, dryRun, a); err != nil {
					log.Error().Err(err).Str("type", "metacritic").Str("client", listsClient.Name).Msg("error while processing Metacritic, continuing with other lists")
					processingErrors = append(processingErrors, fmt.Sprintf("Metacritic - %s: %v", listsClient.Name, err))
				}

			case domain.ListTypePlaintext:
				if err := s.plaintext(ctx, listsClient, dryRun, a); err != nil {
					log.Error().Err(err).Str("type", "plaintext").Str("client", listsClient.Name).Msg("error while processing Plaintext list, continuing with other lists")
					processingErrors = append(processingErrors, fmt.Sprintf("Plaintext - %s: %v", listsClient.Name, err))
				}
			}
		}
	}

	return processingErrors
}

func (s Service) Process(processType string, dryRun bool) error {
	if s.cfg.Clients.Autobrr == nil {
		log.Fatal().Msg("must supply omegabrr configuration!")
		return errors.New("must supply omegabrr configuration")
	}

	a := autobrr.NewClient(s.cfg.Clients.Autobrr.Host, s.cfg.Clients.Autobrr.Apikey)
	if s.cfg.Clients.Autobrr.BasicAuth != nil {
		a.SetBasicAuth(s.cfg.Clients.Autobrr.BasicAuth.User, s.cfg.Clients.Autobrr.BasicAuth.Pass)
	}

	log.Debug().Msgf("starting filter processing...")

	start := time.Now()

	g, ctx := errgroup.WithContext(context.Background())

	var processingErrors []string

	switch processType {
	case "arr":
		processingErrors = append(processingErrors, s.ProcessArrs(ctx, dryRun, a)...)
	case "lists":
		processingErrors = append(processingErrors, s.ProcessLists(ctx, dryRun, a)...)
	case "both":
		processingErrors = append(processingErrors, s.ProcessArrs(ctx, dryRun, a)...)
		processingErrors = append(processingErrors, s.ProcessLists(ctx, dryRun, a)...)
	default:
		log.Error().Msgf("Invalid process type: %s", processType)
		return errors.New("invalid process type")
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
		log.Fatal().Msg("must supply omegabrr configuration!")
		return nil, errors.New("must supply omegabrr configuration")
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
