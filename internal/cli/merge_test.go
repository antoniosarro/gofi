package cli

import (
	"flag"
	"testing"

	"github.com/antoniosarro/gofi/internal/config"
)

func TestMergeWithConfig(t *testing.T) {
	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)

	// Create config with some values
	cfg := &config.Config{
		Modules: map[string]*config.ModuleConfig{
			"application": {
				Enabled:          true,
				EnablePagination: false,
				ItemsPerPage:     8,
				EnableTags:       false,
				EnableHighlight:  false,
				EnableFavorites:  false,
				Settings:         make(map[string]interface{}),
			},
		},
	}

	// Create options with some values
	opts := &Options{
		Module:           "application",
		EnablePagination: true,
		ItemsPerPage:     15,
		EnableTags:       true,
		EnableHighlight:  true,
		EnableFavorites:  true,
	}

	// Simulate flags being set
	flag.Bool("pagination", false, "")
	flag.Int("items-per-page", 8, "")
	flag.Bool("tags", false, "")

	// Set the flags to simulate CLI usage
	flag.Set("pagination", "true")
	flag.Set("items-per-page", "15")
	flag.Set("tags", "true")

	// Merge
	merged := opts.MergeWithConfig(cfg)

	// Check that CLI flags override config
	if !merged.EnablePagination {
		t.Error("EnablePagination should be true (from CLI)")
	}

	if merged.ItemsPerPage != 15 {
		t.Errorf("ItemsPerPage = %d, want 15 (from CLI)", merged.ItemsPerPage)
	}

	if !merged.EnableTags {
		t.Error("EnableTags should be true (from CLI)")
	}

	// Check that non-set flags keep config values
	// (EnableHighlight and EnableFavorites were not set via flag.Set)
	if merged.EnableHighlight {
		t.Error("EnableHighlight should be false (from config, not overridden)")
	}

	if merged.EnableFavorites {
		t.Error("EnableFavorites should be false (from config, not overridden)")
	}
}

func TestMergeWithConfigNoModule(t *testing.T) {
	// Test merging when module doesn't exist in config
	cfg := &config.Config{
		Modules: map[string]*config.ModuleConfig{},
	}

	opts := &Options{
		Module:           "newmodule",
		EnablePagination: true,
	}

	merged := opts.MergeWithConfig(cfg)

	if merged == nil {
		t.Fatal("MergeWithConfig() returned nil")
	}

	if !merged.Enabled {
		t.Error("New module should be enabled by default")
	}
}

func TestMergeWithConfigNoFlags(t *testing.T) {
	// Reset flags
	flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)

	cfg := &config.Config{
		Modules: map[string]*config.ModuleConfig{
			"application": {
				Enabled:          true,
				EnablePagination: true,
				ItemsPerPage:     10,
				EnableTags:       true,
				Settings:         make(map[string]interface{}),
			},
		},
	}

	opts := &Options{
		Module: "application",
	}

	// Don't set any flags
	merged := opts.MergeWithConfig(cfg)

	// All values should come from config
	if !merged.EnablePagination {
		t.Error("EnablePagination should be true (from config)")
	}

	if merged.ItemsPerPage != 10 {
		t.Errorf("ItemsPerPage = %d, want 10 (from config)", merged.ItemsPerPage)
	}

	if !merged.EnableTags {
		t.Error("EnableTags should be true (from config)")
	}
}
