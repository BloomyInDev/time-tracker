package config

import "testing"

// TestLoadDefaults pins the defaults that matter for a deployment nobody
// configured: cookies secure, WAL off, the usual port. The variables are
// blanked first, since every getter reads an empty value as unset and the
// developer running this may well have some of them exported.
func TestLoadDefaults(t *testing.T) {
	for _, key := range []string{
		"TRACKER_PORT", "TRACKER_DB_PATH", "TRACKER_JWT_SECRET",
		"TRACKER_SQLITE_WAL", "TRACKER_SECURE_COOKIES",
	} {
		t.Setenv(key, "")
	}

	cfg := Load()

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if !cfg.SecureCookies {
		t.Error("SecureCookies = false, want true so an unconfigured deploy is the safe one")
	}
	if cfg.SQLiteWAL {
		t.Error("SQLiteWAL = true, want false so no -wal/-shm files appear unasked")
	}
}

func TestGetPort(t *testing.T) {
	tests := []struct {
		name string
		set  string
		want int
	}{
		{"unset falls back", "", 8080},
		{"valid port", "3000", 3000},
		{"lowest valid", "1", 1},
		{"highest valid", "65535", 65535},
		{"zero rejected", "0", 8080},
		{"negative rejected", "-1", 8080},
		{"above range rejected", "65536", 8080},
		{"not a number", "http", 8080},
		{"empty is unset", "", 8080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TRACKER_TEST_PORT", tt.set)
			if got := getPort("TRACKER_TEST_PORT", 8080); got != tt.want {
				t.Errorf("getPort(%q) = %d, want %d", tt.set, got, tt.want)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name     string
		set      string
		fallback bool
		want     bool
	}{
		{"unset keeps default", "", true, true},
		{"true", "true", false, true},
		{"one", "1", false, true},
		{"false", "false", true, false},
		{"zero", "0", true, false},
		{"garbage keeps default", "yes please", true, true},
		{"garbage keeps a false default", "nope", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TRACKER_TEST_BOOL", tt.set)
			if got := getBool("TRACKER_TEST_BOOL", tt.fallback); got != tt.want {
				t.Errorf("getBool(%q, %t) = %t, want %t", tt.set, tt.fallback, got, tt.want)
			}
		})
	}
}
