package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	netHTTP "net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/autobrr/omegabrr/internal/apitoken"
	"github.com/autobrr/omegabrr/internal/buildinfo"
	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/internal/http"
	"github.com/autobrr/omegabrr/internal/processor"
	"github.com/autobrr/omegabrr/internal/scheduler"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/pflag"
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
`

func init() {
	pflag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), usage)
	}
}

func main() {
	var (
		configPath  string
		dryRun      bool
		tokenLength int
	)

	// Define and parse flags using pflag
	pflag.StringVarP(&configPath, "config", "c", "", "path to configuration file")
	pflag.BoolVar(&dryRun, "dry-run", false, "dry-run without inserting filters")
	pflag.IntVar(&tokenLength, "length", 16, "length of the generated API token")

	pflag.Parse()

	if configPath == "" {
		configPath = os.Getenv("OMEGABRR_CONFIG")

		if _, err := os.Stat(configPath); err != nil {
			userConfigDir, err := os.UserConfigDir()
			if err != nil {
				log.Error().Err(err).Msg("failed to get user config directory")
			}

			base := []string{filepath.Join(userConfigDir, "omegabrr"), "/config"}
			configs := []string{"config.yaml", "config.yml"}

			configPath = ""
			for _, b := range base {
				for _, c := range configs {
					p := filepath.Join(b, c)
					if _, err := os.Stat(p); err != nil {
						continue
					}

					configPath = p
					break
				}

				if configPath != "" {
					break
				}
			}
		}
	}

	cfg := domain.NewConfig(configPath)

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	switch cmd := pflag.Arg(0); cmd {
	case "version":
		err := commandVersion()
		if err != nil {
			log.Error().Err(err).Msg("got error from version check")
		}

	case "update":
		commandUpdate()

	case "generate-token":
		commandGenerateToken(tokenLength)

	case "arr":
		commandProcessArrs(cfg, dryRun)

	case "lists":
		commandProcessLists(cfg, dryRun)

	case "run":
		commandRun(cfg)

	default:
		pflag.Usage()
		if cmd != "help" {
			os.Exit(0)
		}
	}
}

func commandRun(cfg *domain.Config) {
	log.Info().Msgf("starting omegabrr, version: %s commit: %s, build date: %s", buildinfo.Version, buildinfo.Commit, buildinfo.Date)
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
		var processingErrors []error

		arrErrors := p.ProcessArrs(ctx, false)
		if len(arrErrors) > 0 {
			processingErrors = append(processingErrors, arrErrors...)
		}

		listErrors := p.ProcessLists(ctx, false)
		if len(listErrors) > 0 {
			processingErrors = append(processingErrors, listErrors...)
		}

		// Print the summary of potential errors
		if len(processingErrors) > 0 {
			log.Warn().Msgf("Run completed with %d errors", len(processingErrors))

			for _, errMsg := range processingErrors {
				log.Error().Err(errMsg).Msg(errMsg.Error())
			}
		} else {
			log.Info().Msg("Run completed successfully without errors")
		}
	}()

	for sig := range sigCh {
		log.Info().Msgf("Received signal: %v", sig)
		schedulerService.Stop()
		os.Exit(0)
	}
}

func commandProcessLists(cfg *domain.Config, dryRun bool) {
	p := processor.NewService(cfg)
	ctx := context.Background()

	processingErrors := p.ProcessLists(ctx, dryRun)
	if len(processingErrors) > 0 {
		log.Warn().Msgf("Run completed with %d errors", len(processingErrors))

		for _, err := range processingErrors {
			log.Error().Err(err).Msg(err.Error())
		}

		os.Exit(1)
	}

	log.Info().Msg("Run completed successfully without errors")
}

func commandProcessArrs(cfg *domain.Config, dryRun bool) {
	p := processor.NewService(cfg)
	ctx := context.Background()

	processingErrors := p.ProcessArrs(ctx, dryRun)
	if len(processingErrors) > 0 {
		log.Warn().Msgf("Run completed with %d errors", len(processingErrors))

		for _, err := range processingErrors {
			log.Error().Err(err).Msg(err.Error())
		}

		os.Exit(1)
	}

	log.Info().Msg("Run completed successfully without errors")
}

func commandGenerateToken(tokenLength int) {
	key, err := apitoken.GenerateToken(tokenLength)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating API token: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "API Token: %v\nCopy and paste into your config file config.yaml\n", key)
}

func commandUpdate() bool {
	v, err := semver.ParseTolerant(buildinfo.Version)
	if err != nil {
		log.Error().Err(err).Msg("could not parse version")
		return true
	}

	latest, err := selfupdate.UpdateSelf(v, "autobrr/omegabrr")
	if err != nil {
		log.Error().Err(err).Msg("Binary update failed")
		return true
	}

	if latest.Version.Equals(v) {
		// latest version is the same as current version. It means current binary is up-to-date.
		log.Info().Msgf("Current binary is the latest version: %s", buildinfo.Version)
	} else {
		log.Info().Msgf("Successfully updated to version: %s", latest.Version)
	}

	return false
}

func commandVersion() error {
	fmt.Printf("Version: %v\nCommit: %v\nBuild date: %v\n", buildinfo.Version, buildinfo.Commit, buildinfo.Date)

	// get the latest release tag from brr-api
	client := &netHTTP.Client{
		Timeout: 10 * time.Second,
	}

	req, err := netHTTP.NewRequestWithContext(context.Background(), netHTTP.MethodGet, "https://api.autobrr.com/repos/autobrr/omegabrr/releases/latest", nil)
	if err != nil {
		return err
	}

	buildinfo.AttachUserAgentHeader(req)

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, netHTTP.ErrHandlerTimeout) {
			return errors.Wrap(err, "Server timed out while fetching latest release from api")
		}

		return errors.Wrap(err, "Failed to fetch latest release from api")
	}
	defer resp.Body.Close()

	// brr-api returns 500 instead of 404 here
	if resp.StatusCode == netHTTP.StatusNotFound || resp.StatusCode == netHTTP.StatusInternalServerError {
		return errors.New("No release found")
	}

	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return errors.Wrap(err, "Failed to decode response from api")
	}

	fmt.Printf("Latest release: %v\n", rel.TagName)

	return nil
}
