package domain

import (
	"bytes"
	"os"
	"strings"
	"text/template"
	"log"

	"github.com/autobrr/omegabrr/internal/apitoken"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/pkg/errors"
)

type BasicAuth struct {
	User string `koanf:"user"`
	Pass string `koanf:"pass"`
}

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
}

func (c *Config) defaults() {
	c.Server.Host = "0.0.0.0"
	c.Server.Port = 7441

	c.Schedule = "0 */6 * * *"

	c.Clients.Autobrr = nil
	c.Clients.Arr = nil
}

var k = koanf.New(".")

func NewConfig(configPath string) *Config {
	cfg := &Config{}

	cfg.defaults()

	if configPath != "" {
		// create config if it doesn't exist
		if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
			if writeErr := cfg.writeFile(configPath); writeErr != nil {
				log.Fatal()
			}
		}

		if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			log.Fatalf("error loading config: %v", err)
		}

		// unmarshal
		if err := k.Unmarshal("", &cfg); err != nil {
			log.Fatal()
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

	token := apitoken.GenerateToken(16)

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
  #    basicAuth:
  #    user: username
  #    pass: password

  arr:
  #  - name: radarr
  #    type: radarr
  #    host: http://localhost:7878
  #    apikey: API_KEY
  #    filters:
  #      - 15 # Change me

  #  - name: radarr4k
  #    type: radarr
  #    host: http://localhost:7878
  #    apikey: API_KEY
  #    filters:
  #      - 16 # Change me

  #  - name: sonarr
  #    type: sonarr
  #    host: http://localhost:8989
  #    apikey: API_KEY
  #    filters:
  #      - 14 # Change me
  #    #excludeAlternateTitles: true # defaults to false
	
  #  - name: readarr
  #    type: readarr
  #    host: http://localhost:8787
  #    apikey: API_KEY
  #    filters:
  #      - 18 # Change me
		
  # - name: lidarr
  #   type: lidarr
  #   host: http://localhost:8686
  #   apikey: API_KEY
  #   filters:
  #     - 32 # Change me

  # - name: whisparr
  #   type: whisparr
  #   host: http://localhost:6969
  #   apikey: API_KEY
  #   filters:
  #     - 69 # Change me`
