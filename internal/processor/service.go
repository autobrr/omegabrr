package processor

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/pkg/autobrr"
)

type Service struct {
	cfg           *domain.Config
	httpClient    *http.Client
	autobrrClient *autobrr.Client
}

func NewService(cfg *domain.Config) *Service {
	s := &Service{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	if cfg != nil {
		s.autobrrClient = s.newAutobrrClient()
	}
	return s
}

func (s Service) newAutobrrClient() *autobrr.Client {
	if s.cfg.Clients.Autobrr == nil {
		log.Fatal().Msg("must supply omegabrr configuration!")
		return nil
	}

	a := autobrr.NewClient(s.cfg.Clients.Autobrr.Host, s.cfg.Clients.Autobrr.Apikey)
	if s.cfg.Clients.Autobrr.BasicAuth != nil {
		a.SetBasicAuth(s.cfg.Clients.Autobrr.BasicAuth.User, s.cfg.Clients.Autobrr.BasicAuth.Pass)
	}

	return a
}

// shouldProcessItem determines if an item should be processed based on its monitored status and configuration
func (s Service) shouldProcessItem(monitored bool, arrConfig *domain.ArrConfig) bool {
	if arrConfig.IncludeUnmonitored {
		return true
	}
	return monitored
}

func (s Service) ProcessArrs(ctx context.Context, dryRun bool) []string {
	var processingErrors []string

	a := s.autobrrClient

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
				if err := s.sonarr(ctx, arrClient, dryRun, a); err != nil {
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

func (s Service) ProcessLists(ctx context.Context, dryRun bool) []string {
	var processingErrors []string

	a := s.autobrrClient

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

			case domain.ListTypeSteam:
				if err := s.steam(ctx, listsClient, dryRun, a); err != nil {
					log.Error().Err(err).Str("type", "steam").Str("client", listsClient.Name).Msg("error while processing Steam wishlist, continuing with other lists")
					processingErrors = append(processingErrors, fmt.Sprintf("Steam - %s: %v", listsClient.Name, err))
				}
			}
		}
	}

	return processingErrors
}

func (s Service) GetFilters(ctx context.Context) ([]autobrr.Filter, error) {
	if s.autobrrClient == nil {
		log.Fatal().Msg("must supply omegabrr configuration!")
		return nil, errors.New("must supply omegabrr configuration")
	}

	filters, err := s.autobrrClient.GetFilters(ctx)
	if err != nil {
		return nil, err
	}

	return filters, nil
}
