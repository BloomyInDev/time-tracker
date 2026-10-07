package i18n

import (
	"context"

	ci "github.com/invopop/ctxi18n/i18n"
)

// DefaultPreset is the vocabulary a user gets when none is chosen. Every
// key a preset may rename is defined under presets.default.
const DefaultPreset = "default"

type presetKey struct{}

// WithPreset attaches the user's vocabulary preset to the context.
func WithPreset(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, presetKey{}, name)
}

func preset(ctx context.Context) string {
	if name, ok := ctx.Value(presetKey{}).(string); ok && name != "" {
		return name
	}
	return DefaultPreset
}

// Vocab translates a key whose wording depends on the user's preset: it
// tries presets.<preset>.<key> and falls back to presets.default.<key>.
func Vocab(ctx context.Context, key string, args ...any) string {
	if name := preset(ctx); name != DefaultPreset {
		if k := "presets." + name + "." + key; ci.Has(ctx, k) {
			return ci.T(ctx, k, args...)
		}
	}
	return ci.T(ctx, "presets."+DefaultPreset+"."+key, args...)
}
