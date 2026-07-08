package config

type Config struct {
	Addr   string
	DBPath string
}

func Load() Config {
	return Config{}
}
