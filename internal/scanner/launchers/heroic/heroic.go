package heroic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/antoniosarro/gofi/internal/domain/entry"
	"github.com/antoniosarro/gofi/internal/scanner/launchers"
)

func init() {
	// Auto-register the Heroic launcher
	launchers.Register(&Launcher{})
}

// Launcher implements the GameLauncher interface for Heroic Games Launcher
type Launcher struct {
	homeDir string
}

// Name returns the launcher identifier
func (l *Launcher) Name() string {
	return "heroic"
}

// Scan discovers games from Heroic Games Launcher
func (l *Launcher) Scan() ([]*entry.Entry, error) {
	l.homeDir = os.Getenv("HOME")
	if l.homeDir == "" {
		return nil, fmt.Errorf("HOME environment variable not set")
	}

	heroicConfigDir := filepath.Join(l.homeDir, ".config", "heroic")
	libraryPath := filepath.Join(heroicConfigDir, "sideload_apps", "library.json")

	// Check if library exists
	if _, err := os.Stat(libraryPath); os.IsNotExist(err) {
		// Not an error - Heroic may not be installed
		return nil, nil
	}

	// Parse library
	library, err := l.parseLibrary(libraryPath)
	if err != nil {
		return nil, fmt.Errorf("parsing heroic library: %w", err)
	}

	// Convert games to entries
	entries := make([]*entry.Entry, 0)
	for _, game := range library.Games {
		// Only process installed sideloaded games
		if !game.IsInstalled || game.Runner != "sideload" {
			continue
		}

		e := l.gameToEntry(game, heroicConfigDir)
		entries = append(entries, e)
	}

	return entries, nil
}

// parseLibrary reads and parses the Heroic library.json
func (l *Launcher) parseLibrary(path string) (*Library, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var library Library
	if err := json.Unmarshal(data, &library); err != nil {
		return nil, err
	}

	return &library, nil
}

// gameToEntry converts a Heroic game to an Entry
func (l *Launcher) gameToEntry(game Game, heroicConfigDir string) *entry.Entry {
	// Try to load categories from GamesConfig
	categories := l.loadGameCategories(heroicConfigDir, game.AppName)

	// If no categories found, use default "Game" category
	if len(categories) == 0 {
		categories = []string{"Game"}
	}

	// Build the heroic:// URI to launch the game
	heroicURI := fmt.Sprintf("heroic://launch/sideload/%s", game.AppName)

	return &entry.Entry{
		Name:       game.Title,
		Comment:    fmt.Sprintf("Heroic Game: %s", game.FolderName),
		Exec:       fmt.Sprintf("xdg-open %s", heroicURI),
		Icon:       "applications-games", // Generic game icon
		Terminal:   false,
		Categories: categories,
		Path:       fmt.Sprintf("heroic-sideload-%s", game.AppName), // Unique identifier
	}
}

// loadGameCategories loads categories from the GamesConfig file
func (l *Launcher) loadGameCategories(heroicConfigDir, appName string) []string {
	configPath := filepath.Join(heroicConfigDir, "GamesConfig", fmt.Sprintf("%s.json", appName))

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	// Parse the JSON - it has dynamic keys
	var rawConfig map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return nil
	}

	// Look for the first key that's not "version" or "explicit"
	for key, value := range rawConfig {
		if key == "version" || key == "explicit" {
			continue
		}

		// Parse the game config
		var gameConfig GameConfig
		if err := json.Unmarshal(value, &gameConfig); err != nil {
			continue
		}

		// Return categories if found
		if len(gameConfig.Categories) > 0 {
			return gameConfig.Categories
		}
	}

	return nil
}
