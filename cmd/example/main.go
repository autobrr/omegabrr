package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/autobrr/go-template/internal/domain"
	"github.com/autobrr/go-template/internal/example"
	"github.com/autobrr/go-template/internal/http"

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

const usage = `example-app

An example description

Usage:
    example-app run               Run example-app
    example-app version           Print version info
    example-app help              Show this help message`

func init() {
	pflag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage)
	}
}

func main() {
	var configPath string
	pflag.StringVar(&configPath, "config", "", "path to configuration file")

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
	//k.Load(env.Provider("EXAMPLE_APP_", ".", func(s string) string {
	//	return strings.Replace(strings.ToLower(
	//		strings.TrimPrefix(s, "EXAMPLE_APP_")), "_", ".", -1)
	//}), nil)

	// setup logger
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	switch cmd := pflag.Arg(0); cmd {
	case "version":
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %v\nCommit: %v\n", version, commit)

	case "run":
		log.Info().Msgf("starting server at http://%v:%v", cfg.Server.Host, cfg.Server.Port)

		exampleService := example.NewService()

		srv := http.NewServer(http.Config{
			Host:           cfg.Server.Host,
			Port:           cfg.Server.Port,
			ExampleService: exampleService,
		})

		errorChannel := make(chan error)
		go func() {
			errorChannel <- srv.Open()
		}()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

		for sig := range sigCh {
			switch sig {
			case syscall.SIGHUP:
				log.Log().Msg("shutting down server sighup")
				os.Exit(1)
			case syscall.SIGINT, syscall.SIGQUIT:
				os.Exit(0)
			case syscall.SIGKILL, syscall.SIGTERM:
				os.Exit(0)
			}
		}
	default:
		pflag.Usage()
		if cmd != "help" {
			os.Exit(1)
		}
	}

}
