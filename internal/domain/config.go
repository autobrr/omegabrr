package domain

type BasicAuth struct {
	User string `koanf:"user"`
	Pass string `koanf:"pass"`
}

type ArrConfig struct {
	Host      string     `koanf:"host"`
	Apikey    string     `koanf:"apikey"`
	BasicAuth *BasicAuth `koanf:"basicauth"`
	Filters   []int      `koanf:"filters"`
}

type AutobrrConfig struct {
	Host      string     `koanf:"host"`
	Apikey    string     `koanf:"apikey"`
	BasicAuth *BasicAuth `koanf:"basicauth"`
}

type Config struct {
	Server struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	} `koanf:"server"`
	Clients struct {
		Autobrr *AutobrrConfig `koanf:"autobrr"`
		Radarr  *ArrConfig     `koanf:"radarr"`
		Sonarr  *ArrConfig     `koanf:"sonarr"`
	} `koanf:"clients"`
}
