// Package config reads runtime settings from the environment. Every variable
// is namespaced with TRACKER_, so the app can't pick up a generic name another
// process on the same host happens to export.
package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port      int
	DBPath    string
	JWTSecret string
	// SQLiteWAL turns on write-ahead logging. It trades two extra files
	// next to the database for readers and writers that no longer block
	// each other.
	SQLiteWAL bool
	// SecureCookies marks cookies Secure, so a browser only ever sends
	// them back over HTTPS. It defaults to on: the app is normally reached
	// through a TLS-terminating reverse proxy, and a deployment that never
	// thought about this should land on the safe setting. Turn it off to
	// serve plain HTTP, or the session cookie gets set and never returned.
	SecureCookies bool
}

func Load() Config {
	return Config{
		Port:          getPort("TRACKER_PORT", 8080),
		DBPath:        getEnv("TRACKER_DB_PATH", "time-tracker.db"),
		JWTSecret:     getEnv("TRACKER_JWT_SECRET", "dev-secret-change-me"),
		SQLiteWAL:     getBool("TRACKER_SQLITE_WAL", false),
		SecureCookies: getBool("TRACKER_SECURE_COOKIES", true),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getPort reads a TCP port. Zero is rejected along with the out-of-range
// values: asking the kernel to pick a free port is never what a server someone
// has to reach was meant to do.
func getPort(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		log.Printf("%s: %q isn't a number, using %d", key, raw, fallback)
		return fallback
	}
	if parsed < 1 || parsed > 65535 {
		log.Printf("%s: %d isn't a valid port, using %d", key, parsed, fallback)
		return fallback
	}
	return parsed
}

// getBool reads a boolean in any form strconv accepts ("1", "true", "off").
// An unparseable value is a typo in the deployment, not a reason to run with a
// setting nobody chose, so it is reported and the default stands.
func getBool(key string, fallback bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		log.Printf("%s: %q isn't a boolean, using %t", key, raw, fallback)
		return fallback
	}
	return parsed
}
