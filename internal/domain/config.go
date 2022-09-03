package domain

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/rs/zerolog/log"
)

type BasicAuth struct {
	User string `koanf:"user"`
	Pass string `koanf:"pass"`
}

type ArrConfig struct {
	Name      string     `koanf:"name"`
	Host      string     `koanf:"host"`
	Apikey    string     `koanf:"apikey"`
	BasicAuth *BasicAuth `koanf:"basicAuth"`
	Filters   []int      `koanf:"filters"`
}

type AutobrrConfig struct {
	Host      string     `koanf:"host"`
	Apikey    string     `koanf:"apikey"`
	BasicAuth *BasicAuth `koanf:"basicAuth"`
}

type Config struct {
	Server struct {
		Host     string `koanf:"host"`
		Port     int    `koanf:"port"`
		APIToken string `koanf:"apiToken"`
	} `koanf:"server"`
	Schedule string `koanf:"schedule"`
	Clients  struct {
		Autobrr *AutobrrConfig `koanf:"autobrr"`
		Radarr  []*ArrConfig   `koanf:"radarr"`
		Sonarr  []*ArrConfig   `koanf:"sonarr"`
	} `koanf:"clients"`
}

func (c *Config) defaults() {
	c.Server.Host = "0.0.0.0"
	c.Server.Port = 7441

	c.Schedule = "* */6 * * *"

	c.Clients.Autobrr = nil
	c.Clients.Sonarr = nil
	c.Clients.Radarr = nil
}

var k = koanf.New(".")

func NewConfig(configPath string) *Config {
	cfg := &Config{}

	cfg.defaults()

	if configPath != "" {
		if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			log.Fatal()
		}

		// unmarshal
		if err := k.Unmarshal("", &cfg); err != nil {
			log.Fatal()
		}
	}

	return cfg
}
