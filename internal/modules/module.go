package modules

import (
	"github.com/antoniosarro/gofi/internal/config"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Module represents a gofi module (application launcher, screenshot, powermenu, etc.)
// Modules are the primary extension mechanism for gofi functionality.
type Module interface {
	// Name returns the module identifier (used in CLI: -m application)
	Name() string

	// Description returns a human-readable description shown in --help
	Description() string

	// Initialize sets up the module with configuration.
	// This is called once before CreateWindow.
	// Modules should perform any necessary setup here (scanning, loading data, etc.)
	Initialize(cfg *config.ModuleConfig) error

	// CreateWindow creates and returns the GTK window for this module.
	// This is called after Initialize, when the GTK application is ready.
	CreateWindow(app *gtk.Application) (Window, error)

	// Cleanup performs any necessary cleanup before shutdown.
	// This is called when the application exits or the module is unloaded.
	Cleanup() error
}

// Window represents a module's window interface.
// This abstraction allows different window implementations while maintaining
// a consistent interface for the application lifecycle.
type Window interface {
	// Show displays the window
	Show()

	// Shutdown performs cleanup specific to this window
	Shutdown()

	// Widget returns the underlying GTK ApplicationWindow
	// This is needed for GTK lifecycle management
	Widget() *gtk.ApplicationWindow
}
