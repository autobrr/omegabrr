package scheduler

import (
	"time"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/internal/processor"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

type Service struct {
	cfg              *domain.Config
	processorService *processor.Service

	cron *cron.Cron
	jobs map[string]cron.EntryID
}

func NewService(cfg *domain.Config, processorSvc *processor.Service) *Service {
	return &Service{
		cfg:              cfg,
		processorService: processorSvc,
		cron: cron.New(cron.WithChain(
			cron.Recover(cron.DefaultLogger),
		)),
		jobs: map[string]cron.EntryID{},
	}
}

func (s *Service) Start() {
	log.Info().Msg("starting scheduler")

	s.cron.Start()

	s.initJobs()

	return
}

func (s *Service) Stop() {
	log.Info().Msg("stopping scheduler")

	s.cron.Stop()

	return
}

func (s *Service) AddJob(job cron.Job, interval string, identifier string) (int, error) {
	if interval == "" {
		interval = "* */6 * * *"
	}

	if id, err := s.cron.AddJob(interval, cron.NewChain(
		cron.SkipIfStillRunning(cron.DiscardLogger)).Then(job),
	); err != nil {
		return 0, err
	} else {
		s.jobs[identifier] = id
	}

	log.Info().Msgf("job successfully added: %v", identifier)

	return 0, nil
}

func (s *Service) initJobs() {
	log.Info().Msg("init jobs")

	time.Sleep(2 * time.Second)

	p := &RunProcessorJob{
		Name:             "process-filters",
		Log:              log.With().Str("job", "process-filters").Logger(),
		ProcessorService: s.processorService,
	}

	if _, err := s.AddJob(p, s.cfg.Schedule, "process-filters"); err != nil {
		log.Error().Err(err).Msg("error adding job: process-filters")
	}

	return
}
