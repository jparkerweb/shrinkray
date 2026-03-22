package presets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const customPresetsFile = "custom_presets.yaml"

// customPresetYAML is the YAML representation of a custom preset.
// Fields mirror the Preset struct with yaml tags.
type customPresetYAML struct {
	Key           string   `yaml:"key"`
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description,omitempty"`
	Category      string   `yaml:"category,omitempty"`
	Codec         string   `yaml:"codec"`
	Container     string   `yaml:"container,omitempty"`
	CRF           int      `yaml:"crf,omitempty"`
	MaxBitrate    string   `yaml:"maxBitrate,omitempty"`
	AudioCodec    string   `yaml:"audioCodec,omitempty"`
	AudioBitrate  string   `yaml:"audioBitrate,omitempty"`
	AudioChannels int      `yaml:"audioChannels,omitempty"`
	Resolution    string   `yaml:"resolution,omitempty"`
	MaxFPS        int      `yaml:"maxFps,omitempty"`
	TargetSizeMB  float64  `yaml:"targetSizeMb,omitempty"`
	TwoPass       bool     `yaml:"twoPass,omitempty"`
	SpeedPreset   string   `yaml:"speedPreset,omitempty"`
	ExtraArgs     []string `yaml:"extraArgs,omitempty"`
	Tags          []string `yaml:"tags,omitempty"`
	Icon          string   `yaml:"icon,omitempty"`
	Custom        bool     `yaml:"custom"`
}

func toYAML(p Preset) customPresetYAML {
	return customPresetYAML{
		Key:           p.Key,
		Name:          p.Name,
		Description:   p.Description,
		Category:      string(p.Category),
		Codec:         p.Codec,
		Container:     p.Container,
		CRF:           p.CRF,
		MaxBitrate:    p.MaxBitrate,
		AudioCodec:    p.AudioCodec,
		AudioBitrate:  p.AudioBitrate,
		AudioChannels: p.AudioChannels,
		Resolution:    p.Resolution,
		MaxFPS:        p.MaxFPS,
		TargetSizeMB:  p.TargetSizeMB,
		TwoPass:       p.TwoPass,
		SpeedPreset:   p.SpeedPreset,
		ExtraArgs:     p.ExtraArgs,
		Tags:          p.Tags,
		Icon:          p.Icon,
		Custom:        true,
	}
}

func fromYAML(y customPresetYAML) Preset {
	cat := Category(y.Category)
	if cat == "" {
		cat = CategoryQuality
	}
	container := y.Container
	if container == "" {
		container = "mp4"
	}
	icon := y.Icon
	if icon == "" {
		icon = "\U0001f527" // wrench
	}

	return Preset{
		Key:           y.Key,
		Name:          y.Name,
		Description:   y.Description,
		Category:      cat,
		Codec:         y.Codec,
		Container:     container,
		CRF:           y.CRF,
		MaxBitrate:    y.MaxBitrate,
		AudioCodec:    y.AudioCodec,
		AudioBitrate:  y.AudioBitrate,
		AudioChannels: y.AudioChannels,
		Resolution:    y.Resolution,
		MaxFPS:        y.MaxFPS,
		TargetSizeMB:  y.TargetSizeMB,
		TwoPass:       y.TwoPass,
		SpeedPreset:   y.SpeedPreset,
		ExtraArgs:     y.ExtraArgs,
		Tags:          y.Tags,
		Icon:          icon,
	}
}

// LoadCustomPresets reads custom presets from the config directory and registers
// them in the global registry. Returns the loaded presets.
func LoadCustomPresets(configDir string) ([]Preset, error) {
	path := filepath.Join(configDir, customPresetsFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no custom presets file — not an error
		}
		return nil, fmt.Errorf("failed to read custom presets: %w", err)
	}

	var items []customPresetYAML
	if err := yaml.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("failed to parse custom presets: %w", err)
	}

	var result []Preset
	for _, item := range items {
		p := fromYAML(item)
		if err := validateCustomPreset(p); err != nil {
			return nil, fmt.Errorf("invalid custom preset %q: %w", p.Key, err)
		}
		Register(p)
		result = append(result, p)
	}

	return result, nil
}

// SaveCustomPreset saves a custom preset to the config directory.
// If the preset key already exists in the custom file, it is updated.
func SaveCustomPreset(configDir string, preset Preset) error {
	if err := validateCustomPreset(preset); err != nil {
		return err
	}

	// Check for conflict with built-in presets
	if isBuiltinKey(preset.Key) {
		return fmt.Errorf("preset key %q conflicts with a built-in preset", preset.Key)
	}

	path := filepath.Join(configDir, customPresetsFile)

	// Load existing custom presets
	var items []customPresetYAML
	data, err := os.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(data, &items); err != nil {
			return fmt.Errorf("failed to parse existing custom presets: %w", err)
		}
	}

	// Update or append
	found := false
	newItem := toYAML(preset)
	for i, item := range items {
		if item.Key == preset.Key {
			items[i] = newItem
			found = true
			break
		}
	}
	if !found {
		items = append(items, newItem)
	}

	// Write back
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	header := "# shrinkray custom presets\n# Each entry defines a custom encoding preset.\n# Required fields: key, name, codec, and either crf or targetSizeMb.\n#\n# Example:\n#   - key: my-preset\n#     name: My Custom Preset\n#     codec: h264\n#     crf: 24\n#     container: mp4\n#     audioCodec: aac\n#     audioBitrate: 128k\n#     custom: true\n\n"

	out, err := yaml.Marshal(items)
	if err != nil {
		return fmt.Errorf("failed to marshal custom presets: %w", err)
	}

	content := header + string(out)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write custom presets: %w", err)
	}

	return nil
}

// validateCustomPreset checks that a custom preset has all required fields.
func validateCustomPreset(p Preset) error {
	if strings.TrimSpace(p.Key) == "" {
		return fmt.Errorf("preset key is required")
	}
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("preset name is required")
	}
	if strings.TrimSpace(p.Codec) == "" {
		return fmt.Errorf("preset codec is required")
	}
	if p.CRF == 0 && p.TargetSizeMB == 0 {
		return fmt.Errorf("preset must have either crf or targetSizeMb set")
	}
	return nil
}

// builtinKeys is the set of keys reserved for built-in presets.
var builtinKeys = map[string]bool{
	"lossless": true, "ultra": true, "high": true,
	"balanced": true, "compact": true, "tiny": true,
	"web": true, "email": true, "archive": true,
	"slideshow": true, "4k-to-1080": true,
	"discord": true, "discord-nitro": true, "whatsapp": true,
	"twitter": true, "instagram": true, "tiktok": true, "youtube": true,
}

// isBuiltinKey returns true if the given key is a built-in preset key.
func isBuiltinKey(key string) bool {
	return builtinKeys[strings.ToLower(key)]
}
