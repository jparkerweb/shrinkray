package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig_NotNil(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
}

func TestDefaultConfig_HasExpectedDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Defaults.Preset != "balanced" {
		t.Errorf("expected default preset 'balanced', got '%s'", cfg.Defaults.Preset)
	}
	if cfg.Defaults.Codec != "h265" {
		t.Errorf("expected default codec 'h265', got '%s'", cfg.Defaults.Codec)
	}
	if cfg.Defaults.CRF != 23 {
		t.Errorf("expected default CRF 23, got %d", cfg.Defaults.CRF)
	}
	if cfg.Output.Suffix != "_shrunk" {
		t.Errorf("expected default suffix '_shrunk', got '%s'", cfg.Output.Suffix)
	}
	if cfg.Output.Mode != "suffix" {
		t.Errorf("expected default output mode 'suffix', got '%s'", cfg.Output.Mode)
	}
	if cfg.Batch.Jobs != 2 {
		t.Errorf("expected default batch jobs 2, got %d", cfg.Batch.Jobs)
	}
	if !cfg.UI.Animations {
		t.Error("expected animations enabled by default")
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("expected no error for nonexistent file, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected default config for nonexistent file")
	}
	if cfg.Defaults.Preset != "balanced" {
		t.Errorf("expected default preset for missing file, got '%s'", cfg.Defaults.Preset)
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	yaml := `defaults:
  preset: compact
  codec: h264
  crf: 28
output:
  suffix: _small
batch:
  jobs: 4
`
	if err := os.WriteFile(cfgPath, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Defaults.Preset != "compact" {
		t.Errorf("expected preset 'compact', got '%s'", cfg.Defaults.Preset)
	}
	if cfg.Defaults.CRF != 28 {
		t.Errorf("expected CRF 28, got %d", cfg.Defaults.CRF)
	}
	if cfg.Output.Suffix != "_small" {
		t.Errorf("expected suffix '_small', got '%s'", cfg.Output.Suffix)
	}
	if cfg.Batch.Jobs != 4 {
		t.Errorf("expected batch jobs 4, got %d", cfg.Batch.Jobs)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(cfgPath, []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	// Loading with empty path should return defaults (since default config file probably doesn't exist)
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestSave_And_Load(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	cfg := DefaultConfig()
	cfg.Defaults.Preset = "ultra"
	cfg.path = cfgPath

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load after Save failed: %v", err)
	}

	if loaded.Defaults.Preset != "ultra" {
		t.Errorf("expected preset 'ultra' after round-trip, got '%s'", loaded.Defaults.Preset)
	}
}

func TestMerge_Overrides(t *testing.T) {
	cfg := DefaultConfig()

	crf := 18
	cfg.Merge(ConfigOverrides{
		Preset: "compact",
		CRF:    &crf,
		Suffix: "_small",
	})

	if cfg.Defaults.Preset != "compact" {
		t.Errorf("expected preset 'compact' after merge, got '%s'", cfg.Defaults.Preset)
	}
	if cfg.Defaults.CRF != 18 {
		t.Errorf("expected CRF 18 after merge, got %d", cfg.Defaults.CRF)
	}
	if cfg.Output.Suffix != "_small" {
		t.Errorf("expected suffix '_small' after merge, got '%s'", cfg.Output.Suffix)
	}
}

func TestMerge_EmptyOverridesNoChange(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.Defaults.Preset

	cfg.Merge(ConfigOverrides{})

	if cfg.Defaults.Preset != original {
		t.Errorf("empty merge should not change preset")
	}
}

func TestConfigDir_WithEnvOverride(t *testing.T) {
	t.Setenv("SHRINKRAY_CONFIG", "/custom/config/dir")

	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir != "/custom/config/dir" {
		t.Errorf("expected /custom/config/dir, got %s", dir)
	}
}

func TestConfigDir_Default(t *testing.T) {
	t.Setenv("SHRINKRAY_CONFIG", "")

	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty config dir")
	}
}

func TestCacheDir(t *testing.T) {
	dir, err := CacheDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty cache dir")
	}
}

func TestLogDir(t *testing.T) {
	dir, err := LogDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty log dir")
	}
}
