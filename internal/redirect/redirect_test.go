package redirect

import "testing"

func TestSanitize(t *testing.T) {
	tests := []struct {
		name string
		dest string
		want string
	}{
		{"empty falls back", "", "/fallback"},
		{"plain path", "/tasks", "/tasks"},
		{"keeps query", "/tasks?client=3", "/tasks?client=3"},
		{"keeps fragment", "/tasks#task-7", "/tasks#task-7"},
		{"absolute url rejected", "https://evil.example/x", "/fallback"},
		{"scheme relative rejected", "//evil.example/x", "/fallback"},
		{"backslash trick rejected", `/\evil.example`, "/fallback"},
		{"relative path rejected", "tasks", "/fallback"},
		{"javascript url rejected", "javascript:alert(1)", "/fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sanitize(tt.dest, "/fallback"); got != tt.want {
				t.Errorf("Sanitize(%q) = %q, want %q", tt.dest, got, tt.want)
			}
		})
	}
}

func TestURLWith(t *testing.T) {
	tests := []struct {
		name string
		base string
		dest string
		want string
	}{
		{"appends", "/login", "/tasks", "/login?redirect=%2Ftasks"},
		{"escapes query", "/lang/fr", "/tasks?client=3", "/lang/fr?redirect=%2Ftasks%3Fclient%3D3"},
		{"existing query", "/login?foo=1", "/tasks", "/login?foo=1&redirect=%2Ftasks"},
		{"unsafe dest dropped", "/login", "https://evil.example", "/login"},
		{"empty dest dropped", "/login", "", "/login"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := URLWith(tt.base, tt.dest); got != tt.want {
				t.Errorf("URLWith(%q, %q) = %q, want %q", tt.base, tt.dest, got, tt.want)
			}
		})
	}
}
