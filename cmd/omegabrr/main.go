package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	netHTTP "net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/autobrr/omegabrr/internal/apitoken"
	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/internal/http"
	"github.com/autobrr/omegabrr/internal/processor"
	"github.com/autobrr/omegabrr/internal/scheduler"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/pflag"
)

var (
	version = "dev"
	commit  = ""
)

const usage = `omegabrr - Automagically turn your monitored titles from your arrs and lists into autobrr filters.

Usage:
  omegabrr [command] [flags]

Commands:
  arr            Run omegabrr arr once
  lists          Run omegabrr lists once
  run            Run omegabrr service on schedule
  generate-token Generate an API Token (optionally call with --length <number>)
  version        Print version info
  update         Update omegabrr to latest version.
  help           Show this help message

Flags:
  -c, --config <path>  Path to configuration file (default is $OMEGABRR_CONFIG, or config.yaml in the default user config directory)
  --dry-run            Dry-run without inserting filters (default false)
  --length <number>    Length of the generated API token (default 16)

Provide a configuration file using one of the following methods:
1. Use the --config <path> or -c <path> flag.
2. Place a config.yaml file in the default user configuration directory (e.g., ~/.config/omegabrr/).
3. Set the OMEGABRR_CONFIG environment variable.

For more information and examples, visit https://github.com/autobrr/omegabrr
` + "\n"

func init() {
	pflag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), usage)
	}
}

func main() {
	var configPath string
	var dryRun bool

	pflag.StringVarP(&configPath, "config", "c", "", "path to configuration file")
	pflag.BoolVar(&dryRun, "dry-run", false, "dry-run without inserting filters")

	// Define and parse flags using pflag
	length := pflag.Int("length", 16, "length of the generated API token")
	pflag.Parse()

	if configPath == "" {
		configPath = os.Getenv("OMEGABRR_CONFIG")

		if configPath == "" {
			userConfigDir, err := os.UserConfigDir()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get user config directory")
			}
			defaultConfigPath := filepath.Join(userConfigDir, "omegabrr", "config.yaml")

			if _, err := os.Stat(defaultConfigPath); err == nil {
				configPath = defaultConfigPath
			}
		}
	}

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	switch cmd := pflag.Arg(0); cmd {
	case "version":
		fmt.Printf("Version: %v\nCommit: %v\n", version, commit)

		// get the latest release tag from brr-api
		client := &netHTTP.Client{
			Timeout: 10 * time.Second,
		}

		resp, err := client.Get("https://api.autobrr.com/repos/autobrr/omegabrr/releases/latest")
		if err != nil {
			if errors.Is(err, netHTTP.ErrHandlerTimeout) {
				fmt.Println("Server timed out while fetching latest release from api")
			} else {
				fmt.Printf("Failed to fetch latest release from api: %v\n", err)
			}
			os.Exit(1)
		}
		defer resp.Body.Close()

		// brr-api returns 500 instead of 404 here
		if resp.StatusCode == netHTTP.StatusNotFound || resp.StatusCode == netHTTP.StatusInternalServerError {
			fmt.Print("No release found")
			os.Exit(1)
		}

		var rel struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
			fmt.Printf("Failed to decode response from api: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Latest release: %v\n", rel.TagName)

	case "update":
		v, err := semver.ParseTolerant(version)
		if err != nil {
			log.Info().Msgf("could not parse version:", err)
			return
		}

		latest, err := selfupdate.UpdateSelf(v, "autobrr/omegabrr")
		if err != nil {
			log.Info().Msgf("Binary update failed:", err)
			return
		}

		if latest.Version.Equals(v) {
			// latest version is the same as current version. It means current binary is up-to-date.
			log.Info().Msgf("Current binary is the latest version", version)
		} else {
			log.Info().Msgf("Successfully updated to version: ", latest.Version)
		}
		break

	case "generate-token":
		key, err := apitoken.GenerateToken(*length)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error generating API token: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "API Token: %v\nCopy and paste into your config file config.yaml\n", key)
	case "arr":
		cfg := domain.NewConfig(configPath)

		p := processor.NewService(cfg)
		ctx := context.Background()
		errors := p.ProcessArrs(ctx, dryRun)
		if len(errors) == 0 {
			log.Info().Msg("Run complete.")
		} else {
			log.Warn().Msg("Run complete, with errors.")
			log.Warn().Msg("Errors encountered during processing:")
			for _, err := range errors {
				log.Warn().Msg(err)
			}
			os.Exit(1)
		}

	case "lists":
		cfg := domain.NewConfig(configPath)

		p := processor.NewService(cfg)
		ctx := context.Background()
		errors := p.ProcessLists(ctx, dryRun)
		if len(errors) == 0 {
			log.Info().Msg("Run complete.")
		} else {
			log.Warn().Msg("Run complete, with errors.")
			log.Warn().Msg("Errors encountered during processing:")
			for _, err := range errors {
				log.Warn().Msg(err)
			}
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
		signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

		schedulerService.Start()

		go func() {
			log.Debug().Msgf("sleeping 15 seconds before running...")

			time.Sleep(15 * time.Second)

			ctx := context.Background()

			// Store processing errors for ProcessArrs and ProcessLists
			var processingErrors []string

			arrsErrors := p.ProcessArrs(ctx, false)
			if len(arrsErrors) > 0 {
				processingErrors = append(processingErrors, arrsErrors...)
			}

			listsErrors := p.ProcessLists(ctx, false)
			if len(listsErrors) > 0 {
				processingErrors = append(processingErrors, listsErrors...)
			}

			// Print the summary of potential errors
			if len(processingErrors) == 0 {
				log.Info().Msg("Run complete.")
			} else {
				log.Warn().Msg("Run complete, with errors.")
				log.Warn().Msg("Errors encountered during processing:")
				for _, errMsg := range processingErrors {
					log.Warn().Msg(errMsg)
				}
			}

		}()

		for sig := range sigCh {
			log.Info().Msgf("Received signal: %v", sig)
			schedulerService.Stop()
			os.Exit(0)
		}

	default:
		pflag.Usage()
		if cmd != "help" {
			os.Exit(0)
		}
	}
}
