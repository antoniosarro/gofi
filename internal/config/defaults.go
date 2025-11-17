package config

// Default returns the default configuration
func Default() *Config {
	return &Config{
		GlobalCSS: "",
		Modules: map[string]*ModuleConfig{
			"application": defaultApplicationConfig(),
			"screenshot":  defaultScreenshotConfig(),
			"powermenu":   defaultPowermenuConfig(),
			"emoji":       defaultEmojiConfig(),
		},
	}
}

// defaultApplicationConfig returns default config for application module
func defaultApplicationConfig() *ModuleConfig {
	return &ModuleConfig{
		Enabled:           true,
		EnablePagination:  false,
		ItemsPerPage:      8,
		EnableTags:        false,
		EnableHighlight:   false,
		EnableFavorites:   false,
		ScanGameLaunchers: true,
		CustomCSS:         "",
		Settings:          make(map[string]interface{}),
	}
}

// defaultScreenshotConfig returns default config for screenshot module
func defaultScreenshotConfig() *ModuleConfig {
	return &ModuleConfig{
		Enabled:  true,
		Settings: make(map[string]interface{}),
	}
}

// defaultPowermenuConfig returns default config for powermenu module
func defaultPowermenuConfig() *ModuleConfig {
	return &ModuleConfig{
		Enabled:  true,
		Settings: make(map[string]interface{}),
	}
}

func defaultEmojiConfig() *ModuleConfig {
	return &ModuleConfig{
		Enabled:          true,
		EnablePagination: false, // Use grid view instead
		ItemsPerPage:     64,    // Items per page if pagination enabled
		Settings: map[string]interface{}{
			"emoji_file": "~/.config/gofi/all_emojis.txt",
		},
	}
}
