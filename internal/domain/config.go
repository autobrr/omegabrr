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

type ArrConfig struct {
	Name      string     `koanf:"name"`
	Type      ArrType    `koanf:"type"`
	Host      string     `koanf:"host"`
	Apikey    string     `koanf:"apikey"`
	BasicAuth *BasicAuth `koanf:"basicAuth"`
	Filters   []int      `koanf:"filters"`
}

type ArrType string

var (
	ArrTypeRadarr ArrType = "radarr"
	ArrTypeSonarr ArrType = "sonarr"
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
			log.Fatal()
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
		pd.Read(b)
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

var configTemplate = `# config.toml
---
server:
  host: {{ .host }}
  port: 7441
  apiToken: {{ .apiToken }}
schedule: 0 */6 * * *
clients:
  autobrr:
    host: http://localhost:7474
    apikey: API_KEY

  arr:
    - name: radarr
      type: radarr
      host: http://localhost:7878
      apikey: API_KEY
      filters:
        - 15

    - name: radarr4k
      type: radarr
      host: http://localhost:7878
      apikey: API_KEY
      filters:
        - 16

    - name: sonarr
      type: sonarr
      host: http://localhost:8989
      apikey: API_KEY
      filters:
        - 14`
