package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/autobrr/omegabrr/internal/apitoken"
	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/internal/http"
	"github.com/autobrr/omegabrr/internal/processor"
	"github.com/autobrr/omegabrr/internal/scheduler"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/pflag"
)

var (
	version = "dev"
	commit  = ""
)

const usage = `omegabrr

Turn your monitored shows from your arrs into autobrr filters, automagically!

Usage:
    omegabrr generate-token    Generate API Token
    omegabrr arr               Run omegabrr once
    omegabrr run               Run omegabrr service
    omegabrr version           Print version info
    omegabrr help              Show this help message`

func init() {
	pflag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage)
	}
}

func main() {
	var configPath string
	var dryRun bool
	pflag.StringVar(&configPath, "config", "", "path to configuration file")
	pflag.BoolVar(&dryRun, "dry-run", false, "dry-run without inserting filters")

	pflag.Parse()

	// setup logger
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	switch cmd := pflag.Arg(0); cmd {
	case "version":
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %v\nCommit: %v\n", version, commit)
	case "generate-token":
		length := flag.Int("length", 16, "length of the generated API token")
		flag.Parse()
		key, err := apitoken.GenerateToken(*length)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error generating API token: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "API Token: %v\nCopy and paste into your config file config.yaml\n", key)

	case "arr":
		cfg := domain.NewConfig(configPath)

		p := processor.NewService(cfg)
		if err := p.Process(dryRun); err != nil {
			log.Error().Err(err).Msgf("error during processing")
			os.Exit(1)
		}

	case "run":
		cfg := domain.NewConfig(configPath)

		log.Info().Msgf("starting omegabrr: %s", version)
		log.Info().Msgf("running on schedule: %v", cfg.Schedule)

		p := processor.NewService(cfg)

		schedulerService := scheduler.NewService(cfg, p)

		srv := http.NewServer(cfg, p)

		errorChannel := make(chan error)
		go func() {
			errorChannel <- srv.Open()
		}()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

		schedulerService.Start()

		go func() {
			log.Debug().Msgf("sleeping 15 seconds before running...")

			time.Sleep(15 * time.Second)

			if err := p.Process(false); err != nil {
				log.Error().Err(err).Msgf("error during initial processing")
			}
		}()

		for sig := range sigCh {
			switch sig {
			case syscall.SIGHUP:
				log.Log().Msg("shutting down server sighup")
				schedulerService.Stop()
				os.Exit(0)
			case syscall.SIGINT, syscall.SIGQUIT:
				schedulerService.Stop()
				os.Exit(0)
			case syscall.SIGKILL, syscall.SIGTERM:
				schedulerService.Stop()
				os.Exit(0)
			}
		}
	default:
		pflag.Usage()
		if cmd != "help" {
			os.Exit(0)
		}
	}
}
