package cli

import (
	"flag"

	"github.com/antoniosarro/gofi/internal/config"
)

// MergeWithConfig merges CLI options with config file for a specific module
// CLI flags take precedence over config file settings
func (opts *Options) MergeWithConfig(cfg *config.Config) *config.ModuleConfig {
	// Get module config
	moduleConfig, ok := cfg.Modules[opts.Module]
	if !ok {
		moduleConfig = &config.ModuleConfig{
			Enabled:  true,
			Settings: make(map[string]interface{}),
		}
	}

	// Check if flags were explicitly set
	flagSet := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		flagSet[f.Name] = true
	})

	// CLI flags override config file only if explicitly set
	if flagSet["pagination"] {
		moduleConfig.EnablePagination = opts.EnablePagination
	}
	if flagSet["tags"] {
		moduleConfig.EnableTags = opts.EnableTags
	}
	if flagSet["highlight"] {
		moduleConfig.EnableHighlight = opts.EnableHighlight
	}
	if flagSet["favorites"] {
		moduleConfig.EnableFavorites = opts.EnableFavorites
	}
	if flagSet["items-per-page"] {
		moduleConfig.ItemsPerPage = opts.ItemsPerPage
	}

	return moduleConfig
}
