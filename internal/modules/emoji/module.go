package emoji

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/antoniosarro/gofi/internal/config"
	"github.com/antoniosarro/gofi/internal/modules"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func init() {
	modules.Register(&Module{})
}

type Module struct {
	config *config.ModuleConfig
	emojis []Emoji
	window *Window
}

func (m *Module) Name() string {
	return "emoji"
}

func (m *Module) Description() string {
	return "Emoji picker with search and clipboard integration"
}

func (m *Module) Initialize(cfg *config.ModuleConfig) error {
	m.config = cfg

	// Get emoji file path from config or use default
	emojiFilePath := m.getEmojiFilePath()

	// Load emojis
	emojis, err := ParseEmojiFile(emojiFilePath)
	if err != nil {
		return fmt.Errorf("failed to load emoji file: %w", err)
	}

	if len(emojis) == 0 {
		return fmt.Errorf("no emojis found in file: %s", emojiFilePath)
	}

	m.emojis = emojis
	return nil
}

func (m *Module) getEmojiFilePath() string {
	// Check if custom path is set in config
	if path, ok := m.config.Settings["emoji_file"].(string); ok && path != "" {
		return config.ExpandPath(path)
	}

	// Try default locations
	homeDir := os.Getenv("HOME")
	defaultPaths := []string{
		filepath.Join(homeDir, ".config/gofi/all_emojis.txt"),
		filepath.Join(homeDir, ".local/share/gofi/all_emojis.txt"),
		"./all_emojis.txt",
		"/usr/share/gofi/all_emojis.txt",
	}

	for _, path := range defaultPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Return default config path (will fail later if not found)
	return filepath.Join(homeDir, ".config/gofi/all_emojis.txt")
}

func (m *Module) CreateWindow(app *gtk.Application) (modules.Window, error) {
	window := NewWindow(app, m.config, m.emojis)
	m.window = window
	return window, nil
}

func (m *Module) Cleanup() error {
	return nil
}
