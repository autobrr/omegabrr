package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/autobrr/omegabrr/internal/domain"
	"github.com/autobrr/omegabrr/internal/processor"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/pflag"
)

var (
	version = "dev"
	commit  = ""
)

var k = koanf.New(".")

const usage = `omegabrr

An example description

Usage:
    omegabrr run               Run omegabrr
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

	cfg := domain.Config{}

	if configPath != "" {
		if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			log.Fatal()
		}

		// unmarshal
		if err := k.Unmarshal("", &cfg); err != nil {
			log.Fatal()
		}
	}

	// setup logger
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	switch cmd := pflag.Arg(0); cmd {
	case "version":
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %v\nCommit: %v\n", version, commit)
	case "arr":
		p := processor.NewService(cfg)
		if err := p.Process(dryRun); err != nil {
			return
		}

	case "run":
		log.Info().Msgf("starting server at http://%v:%v", cfg.Server.Host, cfg.Server.Port)

		//exampleService := example.NewService()
		//
		//srv := http.NewServer(http.Config{
		//	Host:           cfg.Server.Host,
		//	Port:           cfg.Server.Port,
		//	ExampleService: exampleService,
		//})
		//
		//errorChannel := make(chan error)
		//go func() {
		//	errorChannel <- srv.Open()
		//}()
		//
		//sigCh := make(chan os.Signal, 1)
		//signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
		//
		//for sig := range sigCh {
		//	switch sig {
		//	case syscall.SIGHUP:
		//		log.Log().Msg("shutting down server sighup")
		//		os.Exit(1)
		//	case syscall.SIGINT, syscall.SIGQUIT:
		//		os.Exit(0)
		//	case syscall.SIGKILL, syscall.SIGTERM:
		//		os.Exit(0)
		//	}
		//}
	default:
		pflag.Usage()
		if cmd != "help" {
			os.Exit(1)
		}
	}

}
