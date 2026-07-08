package config

import "os"

type Config struct {
	Port      int
	DBPath    string
	JWTSecret string
}

func Load() Config {
	return Config{
		Port:      8080,
		DBPath:    getEnv("DB_PATH", "time-tracker.db"),
		JWTSecret: getEnv("JWT_SECRET", "dev-secret-change-me"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
