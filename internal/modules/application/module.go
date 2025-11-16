package application

import (
	"github.com/antoniosarro/gofi/internal/config"
	"github.com/antoniosarro/gofi/internal/modules"
	"github.com/antoniosarro/gofi/internal/scanner"
	"github.com/antoniosarro/gofi/internal/ui"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func init() {
	// Auto-register the application module
	modules.Register(&Module{})
}

// Module implements the modules.Module interface for the application launcher
type Module struct {
	scanner *scanner.Scanner
	window  *ui.Window
	config  *config.ModuleConfig
}

// Name returns the module identifier
func (m *Module) Name() string {
	return "application"
}

// Description returns a human-readable description
func (m *Module) Description() string {
	return "Application launcher with fuzzy search and favorites"
}

// Initialize sets up the application launcher module
func (m *Module) Initialize(cfg *config.ModuleConfig) error {
	m.config = cfg

	// Create scanner with configuration
	s, err := scanner.NewScanner(cfg.EnableFavorites, cfg.ScanGameLaunchers)
	if err != nil {
		return err
	}

	// Scan for applications and games
	// This initializes the search engine internally
	if err := s.Scan(); err != nil {
		return err
	}

	m.scanner = s
	return nil
}

// CreateWindow creates the application launcher window
func (m *Module) CreateWindow(app *gtk.Application) (modules.Window, error) {
	window := ui.New(app, m.scanner, m.config)
	m.window = window
	return window, nil
}

// Cleanup performs cleanup before shutdown
func (m *Module) Cleanup() error {
	if m.scanner != nil {
		// Scanner handles favorites saving
		return nil
	}
	return nil
}
