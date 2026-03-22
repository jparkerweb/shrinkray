package presets

import (
	"strings"
)

var registry []Preset

func init() {
	for _, p := range qualityPresets() {
		Register(p)
	}
	for _, p := range purposePresets() {
		Register(p)
	}
	for _, p := range platformPresets() {
		Register(p)
	}
}

// Register adds a preset to the global registry.
func Register(p Preset) {
	registry = append(registry, p)
}

// All returns a copy of all registered presets.
func All() []Preset {
	result := make([]Preset, len(registry))
	copy(result, registry)
	return result
}

// ByCategory returns all presets matching the given category.
func ByCategory(cat Category) []Preset {
	var result []Preset
	for _, p := range registry {
		if p.Category == cat {
			result = append(result, p)
		}
	}
	return result
}

// Lookup finds a preset by query string with the following match priority:
// 1. Exact key match
// 2. Case-insensitive key match
// 3. Tag match
// 4. Fuzzy substring in name or description
func Lookup(query string) (Preset, bool) {
	q := strings.TrimSpace(query)
	if q == "" {
		return Preset{}, false
	}

	// 1. Exact key match
	for _, p := range registry {
		if p.Key == q {
			return p, true
		}
	}

	// 2. Case-insensitive key match
	lower := strings.ToLower(q)
	for _, p := range registry {
		if strings.ToLower(p.Key) == lower {
			return p, true
		}
	}

	// 3. Tag match
	for _, p := range registry {
		for _, tag := range p.Tags {
			if strings.ToLower(tag) == lower {
				return p, true
			}
		}
	}

	// 4. Fuzzy substring in name or description
	for _, p := range registry {
		if strings.Contains(strings.ToLower(p.Name), lower) {
			return p, true
		}
		if strings.Contains(strings.ToLower(p.Description), lower) {
			return p, true
		}
	}

	return Preset{}, false
}
