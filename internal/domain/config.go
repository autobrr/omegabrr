package domain

type Config struct {
	Server struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	} `koanf:"server"`
}
