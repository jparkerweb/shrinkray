package config

import (
	"os"
	"path/filepath"
)

const appName = "shrinkray"

// ConfigDir returns the path to the shrinkray configuration directory.
// It respects the SHRINKRAY_CONFIG environment variable as an override.
func ConfigDir() (string, error) {
	if envDir := os.Getenv("SHRINKRAY_CONFIG"); envDir != "" {
		return envDir, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName), nil
}

// CacheDir returns the path to the shrinkray cache directory.
func CacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName), nil
}

// LogDir returns the path to the shrinkray log directory.
// Logs are stored inside the cache directory.
func LogDir() (string, error) {
	cache, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cache, "logs"), nil
}
