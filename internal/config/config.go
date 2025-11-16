package config

import "github.com/antoniosarro/gofi/internal/config/parser"

// Type aliases for parser types
type tomlTable = parser.TOMLTable

// newTOMLParser creates a new TOML parser
func newTOMLParser() *parser.TOMLParser {
	return parser.New()
}

// Config represents the application configuration
type Config struct {
	GlobalCSS string
	Modules   map[string]*ModuleConfig
}

// ModuleConfig represents configuration for a specific module
type ModuleConfig struct {
	Enabled           bool
	EnablePagination  bool
	ItemsPerPage      int
	EnableTags        bool
	EnableHighlight   bool
	EnableFavorites   bool
	ScanGameLaunchers bool
	CustomCSS         string
	// Module-specific settings stored as generic map
	Settings map[string]interface{}
}

// Load loads configuration from a TOML file
func Load(path string) (*Config, error) {
	config := Default()

	// Check if config file exists
	if !fileExists(path) {
		return config, nil
	}

	// Parse TOML file
	parser := newTOMLParser()
	data, err := parser.ParseFile(path)
	if err != nil {
		return nil, err
	}

	// Apply configuration from file
	if err := applyConfig(config, data); err != nil {
		return nil, err
	}

	return config, nil
}

// applyConfig applies parsed TOML data to config
func applyConfig(config *Config, data tomlTable) error {
	// Apply global CSS
	if globalCSS, ok := data.GetString("global_css"); ok {
		config.GlobalCSS = globalCSS
	}

	// Apply module configurations
	if moduleTable, ok := data.GetTable("module"); ok {
		for moduleName, moduleData := range moduleTable {
			moduleConfig, ok := moduleData.(tomlTable)
			if !ok {
				continue
			}

			mc := config.Modules[moduleName]
			if mc == nil {
				mc = &ModuleConfig{
					Enabled:  true,
					Settings: make(map[string]interface{}),
				}
			}

			// Apply module settings
			applyModuleConfig(mc, moduleConfig)
			config.Modules[moduleName] = mc
		}
	}

	return nil
}

// applyModuleConfig applies module-specific configuration
func applyModuleConfig(mc *ModuleConfig, data tomlTable) {
	if enabled, ok := data.GetBool("enabled"); ok {
		mc.Enabled = enabled
	}
	if enablePag, ok := data.GetBool("enable_pagination"); ok {
		mc.EnablePagination = enablePag
	}
	if itemsPerPage, ok := data.GetInt("items_per_page"); ok {
		mc.ItemsPerPage = itemsPerPage
	}
	if enableTags, ok := data.GetBool("enable_tags"); ok {
		mc.EnableTags = enableTags
	}
	if enableHighlight, ok := data.GetBool("enable_highlight"); ok {
		mc.EnableHighlight = enableHighlight
	}
	if enableFavorites, ok := data.GetBool("enable_favorites"); ok {
		mc.EnableFavorites = enableFavorites
	}
	if scanGames, ok := data.GetBool("scan_game_launchers"); ok {
		mc.ScanGameLaunchers = scanGames
	}
	if customCSS, ok := data.GetString("custom_css"); ok {
		mc.CustomCSS = customCSS
	}

	// Store all settings for module-specific use
	for key, value := range data {
		mc.Settings[key] = value
	}
}
