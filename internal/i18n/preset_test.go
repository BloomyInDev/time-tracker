package i18n

import (
	"context"
	"testing"

	"github.com/invopop/ctxi18n"
)

func TestVocab(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	ctx, err := ctxi18n.WithLocale(context.Background(), "en")
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name, preset, key, want string
	}{
		{"no preset", "", "nav.clients", "Clients"},
		{"default", "default", "nav.clients", "Clients"},
		{"renamed key", "projects", "nav.clients", "Projects"},
		{"unknown preset", "nope", "nav.clients", "Clients"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := ctx
			if tt.preset != "" {
				c = WithPreset(ctx, tt.preset)
			}
			if got := Vocab(c, tt.key); got != tt.want {
				t.Errorf("Vocab(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}
