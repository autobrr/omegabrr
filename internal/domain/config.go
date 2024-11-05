package domain

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/autobrr/omegabrr/internal/apitoken"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type BasicAuth struct {
	User string `koanf:"user"`
	Pass string `koanf:"pass"`
}

type ListConfig struct {
	Name         string            `koanf:"name"`
	Type         ListType          `koanf:"type"`
	URL          string            `koanf:"url"`
	BasicAuth    *BasicAuth        `koanf:"basicAuth"`
	Filters      []int             `koanf:"filters"`
	MatchRelease bool              `koanf:"matchRelease"`
	Album        bool              `koanf:"album"`
	Headers      map[string]string `koanf:"headers"`
}

type ListType string

var (
	ListTypeTrakt      ListType = "trakt"
	ListTypeMdblist    ListType = "mdblist"
	ListTypeMetacritic ListType = "metacritic"
	ListTypePlaintext  ListType = "plaintext"
	ListTypeSteam      ListType = "steam"
)

type ArrConfig struct {
	Name                   string     `koanf:"name"`
	Type                   ArrType    `koanf:"type"`
	Host                   string     `koanf:"host"`
	Apikey                 string     `koanf:"apikey"`
	BasicAuth              *BasicAuth `koanf:"basicAuth"`
	Filters                []int      `koanf:"filters"`
	TagsInclude            []string   `koanf:"tagsInclude"`
	TagsExclude            []string   `koanf:"tagsExclude"`
	MatchRelease           bool       `koanf:"matchRelease"`
	ExcludeAlternateTitles bool       `koanf:"excludeAlternateTitles"`
	IncludeUnmonitored     bool       `koanf:"includeUnmonitored"`
}

type ArrType string

var (
	ArrTypeRadarr   ArrType = "radarr"
	ArrTypeSonarr   ArrType = "sonarr"
	ArrTypeReadarr  ArrType = "readarr"
	ArrTypeLidarr   ArrType = "lidarr"
	ArrTypeWhisparr ArrType = "whisparr"
)

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
		Arr     []*ArrConfig   `koanf:"arr"`
	} `koanf:"clients"`
	Lists []*ListConfig `koanf:"lists"`
}

func (c *Config) defaults() {
	c.Server.Host = "0.0.0.0"
	c.Server.Port = 7441

	c.Schedule = "0 */6 * * *"

	c.Clients.Autobrr = nil
	c.Clients.Arr = nil
	c.Lists = nil
}

var k = koanf.New(".")

func validateConfig(condition bool, field, entityType, entityName string) {
	if condition {
		log.Fatal().
			Str("service", "config").
			Msgf("%s not set for %s: %s", field, entityType, entityName)
		os.Exit(1)
	}
}

func NewConfig(configPath string) *Config {
	cfg := &Config{}

	cfg.defaults()

	if configPath != "" {
		// create config if it doesn't exist
		if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
			if writeErr := cfg.writeFile(configPath); writeErr != nil {
				log.Fatal().
					Err(writeErr).
					Str("service", "config").
					Msgf("failed writing %q", configPath)
			}
		}

		if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			log.Fatal().
				Err(err).
				Str("service", "config").
				Msgf("failed parsing %q", configPath)
		}

		// unmarshal
		if err := k.Unmarshal("", &cfg); err != nil {
			log.Fatal().
				Err(err).
				Str("service", "config").
				Msgf("failed unmarshalling %q", configPath)
		}

		for _, list := range cfg.Lists {
			validateConfig(len(list.Filters) < 1, "Filters", "list", list.Name)
			validateConfig(list.URL == "", "URL", "list", list.Name)
			validateConfig(list.Type == "", "Type", "list", list.Name)
		}

		for _, arr := range cfg.Clients.Arr {
			validateConfig(len(arr.Filters) < 1, "Filters", "arr", arr.Name)
			validateConfig(arr.Host == "", "Host", "arr", arr.Name)
			validateConfig(arr.Apikey == "", "API", "arr", arr.Name)
			validateConfig(arr.Type == "", "Type", "arr", arr.Name)
		}

	}

	return cfg
}

func (c *Config) writeFile(configPath string) error {

	// set default host
	host := "127.0.0.1"

	if _, err := os.Stat("/.dockerenv"); err == nil {
		// docker creates a .dockerenv file at the root
		// of the directory tree inside the container.
		// if this file exists then the viewer is running
		// from inside a container so return true
		host = "0.0.0.0"
	} else if pd, _ := os.Open("/proc/1/cgroup"); pd != nil {
		defer pd.Close()
		b := make([]byte, 4096, 4096)
		_, _ = pd.Read(b)
		if strings.Contains(string(b), "/docker") || strings.Contains(string(b), "/lxc") {
			host = "0.0.0.0"
		}
	}

	f, err := os.Create(configPath)
	if err != nil { // perm 0666
		// handle failed create
		log.Printf("error creating config file: %q", err)
		//return err
	}
	defer f.Close()

	// setup text template to inject variables into
	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return errors.Wrap(err, "could not create config template")
	}

	token, err := apitoken.GenerateToken(16)
	if err != nil {
		return errors.Wrap(err, "Error generating token: %v")
	}

	tmplVars := map[string]string{
		"host":     host,
		"apiToken": token,
	}

	var buffer bytes.Buffer
	if err = tmpl.Execute(&buffer, &tmplVars); err != nil {
		return errors.Wrap(err, "could not write torrent url template output")
	}

	if _, err = f.WriteString(buffer.String()); err != nil {
		log.Printf("error writing contents to file: %v %q", configPath, err)
		return err
	}

	return f.Sync()
}

var configTemplate = `# config.yaml
---
server:
  host: {{ .host }}
  port: 7441
  apiToken: {{ .apiToken }}
schedule: 0 */6 * * *
clients:
  autobrr:
  #  host: http://localhost:7474
  #  apikey: API_KEY
  #  basicAuth:
  #    user: username
  #    pass: password

  arr:
    #- name: radarr
    #  type: radarr
    #  host: http://localhost:7878
    #  apikey: API_KEY
    #  filters:
    #    - 15 # Change me
    #  includeUnmonitored: false # Set to true to include unmonitored items

    #- name: radarr4k
    #  type: radarr
    #  host: http://localhost:7878
    #  apikey: API_KEY
    #  filters:
    #    - 16 # Change me
    #  includeUnmonitored: false # Set to true to include unmonitored items

    #- name: sonarr
    #  type: sonarr
    #  host: http://localhost:8989
    #  apikey: API_KEY
    #  filters:
    #    - 14 # Change me
    #  includeUnmonitored: false # Set to true to include unmonitored items
    #  #excludeAlternateTitles: true # defaults to false

    #- name: readarr
    #  type: readarr
    #  host: http://localhost:8787
    #  apikey: API_KEY
    #  filters:
    #    - 18 # Change me
    #  includeUnmonitored: false # Set to true to include unmonitored items

    #- name: lidarr
    #  type: lidarr
    #  host: http://localhost:8686
    #  apikey: API_KEY
    #  filters:
    #    - 32 # Change me
    #  includeUnmonitored: false # Set to true to include unmonitored items

    #- name: whisparr
    #  type: whisparr
    #  host: http://localhost:6969
    #  apikey: API_KEY
    #  matchRelease: true
    #  filters:
    #    - 69 # Change me
    #  includeUnmonitored: false # Set to true to include unmonitored items

lists:
  #- name: Latest TV Shows
  #  type: mdblist
  #  url: https://mdblist.com/lists/garycrawfordgc/latest-tv-shows/json
  #  filters:
  #    - 1 # Change me

  #- name: Anticipated TV
  #  type: trakt
  #  url: https://api.autobrr.com/lists/trakt/anticipated-tv
  #  filters:
  #    - 22 # Change me

  #- name: Upcoming Movies
  #  type: trakt
  #  url: https://api.autobrr.com/lists/trakt/upcoming-movies
  #  filters:
  #    - 21 # Change me

  #- name: Upcoming Bluray
  #  type: trakt
  #  url: https://api.autobrr.com/lists/trakt/upcoming-bluray
  #  filters:
  #    - 24 # Change me

  #- name: Popular TV
  #  type: trakt
  #  url: https://api.autobrr.com/lists/trakt/popular-tv
  #  filters:
  #    - 25 # Change me

  #- name: StevenLu
  #  type: trakt
  #  url: https://api.autobrr.com/lists/stevenlu
  #  filters:
  #    - 23 # Change me

  #- name: New Albums
  #  type: metacritic
  #  url: https://api.autobrr.com/lists/metacritic/new-albums
  #  filters:
  #    - 9 # Change me

  #- name: Upcoming Albums
  #  type: metacritic
  #  url: https://api.autobrr.com/lists/metacritic/upcoming-albums
  #  filters:
  #    - 20 # Change me

  #- name: Steam Wishlist
  #  type: steam
  #  url: https://store.steampowered.com/wishlist/id/USERNAME/wishlistdata
  #  filters:
  #    - 20 # Change me
`
