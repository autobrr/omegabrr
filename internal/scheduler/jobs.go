package scheduler

import (
	"github.com/autobrr/omegabrr/internal/processor"

	"github.com/rs/zerolog"
)

type RunProcessorJob struct {
	Name             string
	Log              zerolog.Logger
	ProcessorService *processor.Service
}

func (j *RunProcessorJob) Run() {
	if err := j.ProcessorService.Process("both", false); err != nil {
		j.Log.Error().Err(err).Msgf("something went wrong running processor")
	}
}
