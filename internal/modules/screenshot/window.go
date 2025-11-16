package screenshot

import (
	"github.com/antoniosarro/gofi/internal/config"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Window implements modules.Window for the screenshot tool
type Window struct {
	window *gtk.ApplicationWindow
	config *config.ModuleConfig
}

// NewWindow creates a new screenshot window
func NewWindow(app *gtk.Application, cfg *config.ModuleConfig) *Window {
	w := &Window{
		window: gtk.NewApplicationWindow(app),
		config: cfg,
	}

	// TODO: Build screenshot UI
	w.window.SetTitle("Screenshot")
	w.window.SetDefaultSize(300, 200)
	w.window.SetDecorated(false)
	w.window.SetResizable(false)

	// Placeholder content
	box := gtk.NewBox(gtk.OrientationVertical, 10)
	label := gtk.NewLabel("Screenshot Tool (Coming Soon)")
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
