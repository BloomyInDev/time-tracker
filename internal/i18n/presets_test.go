package i18n

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// leaves flattens a nested YAML map into dotted keys.
func leaves(prefix string, m map[string]any, out map[string]bool) {
	for k, v := range m {
		key := prefix + k
		if sub, ok := v.(map[string]any); ok {
			leaves(key+".", sub, out)
		} else {
			out[key] = true
		}
	}
}

func loadKeys(t *testing.T, locale string) map[string]bool {
	t.Helper()
	raw, err := localesFS.ReadFile("locales/" + locale + ".yml")
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	keys := map[string]bool{}
	leaves("", doc[locale], keys)
	return keys
}

// TestPresetsFallBackToDefault fails when a preset renames a key that
// "default" does not define (the fallback would have nothing to land on),
// when a key lives both in the base and in "default", or when en and fr
// define different preset keys.
func TestPresetsFallBackToDefault(t *testing.T) {
	en, fr := loadKeys(t, "en"), loadKeys(t, "fr")

	for key := range en {
		after, isPreset := strings.CutPrefix(key, "presets.")
		if !isPreset {
			continue
		}
		name, rest, _ := strings.Cut(after, ".")
		if !en["presets.default."+rest] {
			t.Errorf("%s has no counterpart in presets.default", key)
		}
		if name == "default" && en[rest] {
			t.Errorf("%s is defined both in the base and in presets.default", rest)
		}
		if !fr[key] {
			t.Errorf("%s is missing in fr.yml", key)
		}
	}
	for key := range fr {
		if strings.HasPrefix(key, "presets.") && !en[key] {
			t.Errorf("%s is missing in en.yml", key)
		}
	}
}
