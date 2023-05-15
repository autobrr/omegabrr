package scheduler

import (
	"context"

	"github.com/autobrr/omegabrr/internal/processor"

	"github.com/rs/zerolog"
)

type RunProcessorJob struct {
	Name             string
	Log              zerolog.Logger
	ProcessorService *processor.Service
}

func (j *RunProcessorJob) Run() {
	ctx := context.Background()

	arrsErrors := j.ProcessorService.ProcessArrs(ctx, false)
	if len(arrsErrors) > 0 {
		j.Log.Error().Msg("Errors encountered during processing Arrs:")
		for _, errMsg := range arrsErrors {
			j.Log.Error().Msg(errMsg)
		}
	}

	listsErrors := j.ProcessorService.ProcessLists(ctx, false)
	if len(listsErrors) > 0 {
		j.Log.Error().Msg("Errors encountered during processing Lists:")
		for _, errMsg := range listsErrors {
			j.Log.Error().Msg(errMsg)
		}
	}
}
