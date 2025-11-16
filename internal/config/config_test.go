package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg == nil {
		t.Fatal("Default() returned nil")
	}

	if cfg.Modules == nil {
		t.Fatal("Default config has nil Modules")
	}

	// Check that default modules exist
	expectedModules := []string{"application", "screenshot", "powermenu"}
	for _, name := range expectedModules {
		if _, ok := cfg.Modules[name]; !ok {
			t.Errorf("Default config missing module: %s", name)
		}
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("Load() error = %v, want nil (should use defaults)", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Should have default config
	if len(cfg.Modules) == 0 {
		t.Error("Load() returned config with no modules")
	}
}

func TestLoadValid(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `global_css = "/path/to/style.css"

[module.application]
enabled = true
enable_pagination = true
items_per_page = 15
enable_tags = true
enable_highlight = true
enable_favorites = true
scan_game_launchers = false
custom_css = "/path/to/custom.css"
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check global CSS
	if cfg.GlobalCSS != "/path/to/style.css" {
		t.Errorf("GlobalCSS = %q, want %q", cfg.GlobalCSS, "/path/to/style.css")
	}

	// Check application module config
	appConfig, ok := cfg.Modules["application"]
	if !ok {
		t.Fatal("application module not found in config")
	}

	if !appConfig.Enabled {
		t.Error("application module should be enabled")
	}

	if !appConfig.EnablePagination {
		t.Error("EnablePagination should be true")
	}

	if appConfig.ItemsPerPage != 15 {
		t.Errorf("ItemsPerPage = %d, want 15", appConfig.ItemsPerPage)
	}

	if !appConfig.EnableTags {
		t.Error("EnableTags should be true")
	}

	if !appConfig.EnableHighlight {
		t.Error("EnableHighlight should be true")
	}

	if !appConfig.EnableFavorites {
		t.Error("EnableFavorites should be true")
	}

	if appConfig.ScanGameLaunchers {
		t.Error("ScanGameLaunchers should be false")
	}

	if appConfig.CustomCSS != "/path/to/custom.css" {
		t.Errorf("CustomCSS = %q, want %q", appConfig.CustomCSS, "/path/to/custom.css")
	}
}

func TestGetConfigPath(t *testing.T) {
	// Save original env
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	originalHOME := os.Getenv("HOME")
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", originalXDG)
		os.Setenv("HOME", originalHOME)
	}()

	// Test with XDG_CONFIG_HOME set
	os.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
	os.Setenv("HOME", "/home/user")

	path := GetConfigPath()
	expected := "/tmp/xdg/gofi/config.toml"
	if path != expected {
		t.Errorf("GetConfigPath() = %q, want %q", path, expected)
	}

	// Test without XDG_CONFIG_HOME
	os.Unsetenv("XDG_CONFIG_HOME")
	path = GetConfigPath()
	expected = "/home/user/.config/gofi/config.toml"
	if path != expected {
		t.Errorf("GetConfigPath() = %q, want %q", path, expected)
	}
}

func TestExpandPath(t *testing.T) {
	originalHOME := os.Getenv("HOME")
	os.Setenv("HOME", "/home/testuser")
	defer os.Setenv("HOME", originalHOME)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Tilde expansion",
			input:    "~/.config/gofi",
			expected: "/home/testuser/.config/gofi",
		},
		{
			name:     "No tilde",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
