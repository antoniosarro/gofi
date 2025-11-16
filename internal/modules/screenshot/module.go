package screenshot

import (
	"github.com/antoniosarro/gofi/internal/config"
	"github.com/antoniosarro/gofi/internal/modules"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func init() {
	modules.Register(&Module{})
}

// Module implements the modules.Module interface for the screenshot tool
type Module struct {
	config *config.ModuleConfig
	window *Window
}

// Name returns the module identifier
func (m *Module) Name() string {
	return "screenshot"
}

// Description returns a human-readable description
func (m *Module) Description() string {
	return "Screenshot tool with area selection and editing"
}

// Initialize sets up the screenshot module
func (m *Module) Initialize(cfg *config.ModuleConfig) error {
	m.config = cfg
	return nil
}

// CreateWindow creates the screenshot window
func (m *Module) CreateWindow(app *gtk.Application) (modules.Window, error) {
	window := NewWindow(app, m.config)
	m.window = window
	return window, nil
}

// Cleanup performs cleanup before shutdown
func (m *Module) Cleanup() error {
	// Cleanup screenshot temp files, etc.
	return nil
}
