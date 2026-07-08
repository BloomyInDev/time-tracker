package config

type Config struct {
	Port   int
	DBPath string
}

func Load() Config {
	return Config{Port: 8080}
}
