package powermenu

import (
	"github.com/antoniosarro/gofi/internal/config"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Window implements modules.Window for the power menu
type Window struct {
	window *gtk.ApplicationWindow
	config *config.ModuleConfig
}

// NewWindow creates a new power menu window
func NewWindow(app *gtk.Application, cfg *config.ModuleConfig) *Window {
	w := &Window{
		window: gtk.NewApplicationWindow(app),
		config: cfg,
	}

	// TODO: Build powermenu UI (buttons for shutdown, reboot, etc.)
	w.window.SetTitle("Power Menu")
	w.window.SetDefaultSize(400, 300)
	w.window.SetDecorated(false)
	w.window.SetResizable(false)

	// Placeholder content
	box := gtk.NewBox(gtk.OrientationVertical, 10)
	label := gtk.NewLabel("Power Menu (Coming Soon)")
	box.Append(label)
	w.window.SetChild(box)

	return w
}

// Show displays the window
func (w *Window) Show() {
	w.window.SetVisible(true)
}

// Shutdown performs cleanup
func (w *Window) Shutdown() {
	// Cleanup if needed
}

// Widget returns the underlying GTK ApplicationWindow
func (w *Window) Widget() *gtk.ApplicationWindow {
	return w.window
}
