package config

import (
	"os"
	"path/filepath"
)

// GetConfigPath returns the default config file path
func GetConfigPath() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "gofi", "config.toml")
	}
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", "gofi", "config.toml")
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ExpandPath expands ~ to home directory in paths
func ExpandPath(path string) string {
	if path == "" {
		return path
	}

	if path[0] == '~' {
		home := os.Getenv("HOME")
		return filepath.Join(home, path[1:])
	}

	return path
}
