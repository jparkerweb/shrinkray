package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all shrinkray configuration.
type Config struct {
	Defaults DefaultsConfig `yaml:"defaults"`
	Output   OutputConfig   `yaml:"output"`
	UI       UIConfig       `yaml:"ui"`
	Batch    BatchConfig    `yaml:"batch"`
	FFmpeg   FFmpegConfig   `yaml:"ffmpeg"`

	// path is the file this config was loaded from (not serialized).
	path string `yaml:"-"`
}

// DefaultsConfig holds default encoding settings.
type DefaultsConfig struct {
	Preset string `yaml:"preset"`
	Codec  string `yaml:"codec"`
	CRF    int    `yaml:"crf"`
}

// OutputConfig holds output file handling settings.
type OutputConfig struct {
	Mode      string `yaml:"mode"`      // suffix, directory, explicit, inplace
	Suffix    string `yaml:"suffix"`    // default "_shrunk"
	Conflict  string `yaml:"conflict"`  // skip, overwrite, autorename
	Directory string `yaml:"directory"` // output directory for directory mode
}

// UIConfig holds user-interface settings.
type UIConfig struct {
	Theme      string `yaml:"theme"`
	Animations bool   `yaml:"animations"`
}

// BatchConfig holds batch processing settings.
type BatchConfig struct {
	Jobs         int    `yaml:"jobs"`
	Sort         string `yaml:"sort"`
	SkipExisting bool   `yaml:"skipExisting"`
}

// FFmpegConfig holds paths to FFmpeg binaries.
type FFmpegConfig struct {
	FFmpegPath  string `yaml:"ffmpegPath"`
	FFprobePath string `yaml:"ffprobePath"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Defaults: DefaultsConfig{
			Preset: "balanced",
			Codec:  "h265",
			CRF:    23,
		},
		Output: OutputConfig{
			Mode:     "suffix",
			Suffix:   "_shrunk",
			Conflict: "autorename",
		},
		UI: UIConfig{
			Theme:      "neon-dusk",
			Animations: true,
		},
		Batch: BatchConfig{
			Jobs:         2,
			Sort:         "name",
			SkipExisting: false,
		},
		FFmpeg: FFmpegConfig{},
	}
}

// Load reads a Config from the given YAML file path.
// If the file does not exist, it returns DefaultConfig with no error.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		dir, err := ConfigDir()
		if err != nil {
			return cfg, nil
		}
		path = filepath.Join(dir, "config.yaml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	cfg.path = path
	return cfg, nil
}

// Save writes the Config to disk at its loaded path (or the default path).
func (c *Config) Save() error {
	path := c.path
	if path == "" {
		dir, err := ConfigDir()
		if err != nil {
			return err
		}
		path = filepath.Join(dir, "config.yaml")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// Merge applies CLI flag overrides onto this Config.
// Only non-zero/non-empty values in the overrides are applied.
func (c *Config) Merge(overrides ConfigOverrides) {
	if overrides.Preset != "" {
		c.Defaults.Preset = overrides.Preset
	}
	if overrides.Codec != "" {
		c.Defaults.Codec = overrides.Codec
	}
	if overrides.CRF != nil {
		c.Defaults.CRF = *overrides.CRF
	}
	if overrides.Suffix != "" {
		c.Output.Suffix = overrides.Suffix
	}
	if overrides.OutputDir != "" {
		c.Output.Directory = overrides.OutputDir
		c.Output.Mode = "directory"
	}
}

// ConfigOverrides holds CLI flag values that override config file settings.
type ConfigOverrides struct {
	Preset    string
	Codec     string
	CRF       *int
	Suffix    string
	OutputDir string
}
