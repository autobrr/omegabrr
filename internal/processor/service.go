package processor

import (
	"context"
	"net/http"
	"time"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/pkg/autobrr"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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

func (s *Service) newAutobrrClient() *autobrr.Client {
	if s.cfg.Clients.Autobrr == nil {
		log.Fatal().Msg("must supply autobrr client configuration!")
		return nil
	}

	client := autobrr.NewClient(s.cfg.Clients.Autobrr.Host, s.cfg.Clients.Autobrr.Apikey)
	if s.cfg.Clients.Autobrr.BasicAuth != nil {
		client.SetBasicAuth(s.cfg.Clients.Autobrr.BasicAuth.User, s.cfg.Clients.Autobrr.BasicAuth.Pass)
	}

	return client
}

func (s *Service) ProcessArrs(ctx context.Context, dryRun bool) []error {
	if s.cfg.Clients.Arr == nil {
		return nil
	}

	var processingErrors []error

	for _, arrClient := range s.cfg.Clients.Arr {
		arrClient := arrClient

		log.Debug().Msgf("run processing for %s - %s", arrClient.Type, arrClient.Name)

		var err error

		switch arrClient.Type {
		case domain.ArrTypeRadarr:
			err = s.radarr(ctx, arrClient, dryRun, s.autobrrClient)

		case domain.ArrTypeWhisparr:
			err = s.sonarr(ctx, arrClient, dryRun, s.autobrrClient)

		case domain.ArrTypeSonarr:
			err = s.sonarr(ctx, arrClient, dryRun, s.autobrrClient)

		case domain.ArrTypeReadarr:
			err = s.readarr(ctx, arrClient, dryRun, s.autobrrClient)

		case domain.ArrTypeLidarr:
			err = s.lidarr(ctx, arrClient, dryRun, s.autobrrClient)

		default:
			err = errors.Errorf("unsupported arr client type: %s", arrClient.Type)
		}

		if err != nil {
			log.Error().Err(err).Str("type", string(arrClient.Type)).Str("client", arrClient.Name).Msgf("error while processing %s, continuing with other clients", arrClient.Type)

			processingErrors = append(processingErrors, errors.Wrapf(err, "%s - %s", arrClient.Type, arrClient.Name))
		}
	}

	return processingErrors

}

func (s *Service) ProcessLists(ctx context.Context, dryRun bool) []error {
	if s.cfg.Lists == nil {
		return nil
	}

	var processingErrors []error

	for _, listsClient := range s.cfg.Lists {
		listsClient := listsClient

		log.Debug().Msgf("run processing for list %s - %s", listsClient.Type, listsClient.Name)

		var err error

		switch listsClient.Type {
		case domain.ListTypeTrakt:
			err = s.trakt(ctx, listsClient, dryRun, s.autobrrClient)

		case domain.ListTypeMdblist:
			err = s.mdblist(ctx, listsClient, dryRun, s.autobrrClient)

		case domain.ListTypeMetacritic:
			err = s.metacritic(ctx, listsClient, dryRun, s.autobrrClient)

		case domain.ListTypePlaintext:
			err = s.plaintext(ctx, listsClient, dryRun, s.autobrrClient)
		}

		if err != nil {
			log.Error().Err(err).Str("type", string(listsClient.Type)).Str("client", listsClient.Name).Msgf("error while processing %s list, continuing with other lists", listsClient.Type)

			processingErrors = append(processingErrors, errors.Wrapf(err, "%s - %s", listsClient.Type, listsClient.Name))
		}
	}

	return processingErrors

}

func (s *Service) GetFilters(ctx context.Context) ([]autobrr.Filter, error) {
	if s.autobrrClient == nil {
		return nil, errors.New("no autobrr client found in config")
	}

	filters, err := s.autobrrClient.GetFilters(ctx)
	if err != nil {
		return nil, err
	}

	return filters, nil
}
