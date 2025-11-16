package ui

import (
	"log"

	"github.com/antoniosarro/gofi/internal/config"
	"github.com/antoniosarro/gofi/internal/domain/entry"
	"github.com/antoniosarro/gofi/internal/scanner"
	"github.com/antoniosarro/gofi/internal/ui/list"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const (
	WindowWidth      = 600
	BaseWindowHeight = 100 // Base height for search box and padding
	RowHeight        = 60  // Approximate height per row
	PageLabelHeight  = 30  // Height for pagination label
	SearchDebounceMs = 150 // Debounce delay in milliseconds
	MaxVisibleRows   = 8   // Maximum rows to show in window when pagination is disabled
)

// Window represents the main launcher window
type Window struct {
	window        *gtk.ApplicationWindow
	searchEntry   *gtk.SearchEntry
	listView      *list.View
	scrolled      *gtk.ScrolledWindow
	pageLabel     *gtk.Label
	scanner       *scanner.Scanner
	moduleConfig  *config.ModuleConfig
	itemsPerPage  int
	debounceTimer glib.SourceHandle
}

// New creates a new launcher window
func New(app *gtk.Application, s *scanner.Scanner, moduleConfig *config.ModuleConfig) *Window {
	itemsPerPage := moduleConfig.ItemsPerPage
	if !moduleConfig.EnablePagination {
		itemsPerPage = 1000 // Large number to show all items (with scroll)
	}

	w := &Window{
		scanner:      s,
		moduleConfig: moduleConfig,
		itemsPerPage: itemsPerPage,
	}

	// Create main application window
	w.window = gtk.NewApplicationWindow(app)
	w.window.SetTitle("Launcher")

	// Calculate dynamic window height based on items per page
	windowHeight := w.calculateWindowHeight()
	w.window.SetDefaultSize(WindowWidth, windowHeight)
	w.window.SetDecorated(false)
	w.window.SetResizable(false)

	// Create main container
	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginTop(10)
	box.SetMarginBottom(10)
	box.SetMarginStart(10)
	box.SetMarginEnd(10)

	// Create search entry
	w.searchEntry = gtk.NewSearchEntry()
	w.searchEntry.SetPlaceholderText("Type to search applications...")
	w.searchEntry.ConnectSearchChanged(w.onSearchChangedDebounced)
	box.Append(w.searchEntry)

	// Create scrolled window for list
	w.scrolled = gtk.NewScrolledWindow()
	w.scrolled.SetVExpand(true)

	// Enable vertical scrolling when pagination is disabled
	if moduleConfig.EnablePagination {
		w.scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	} else {
		w.scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAlways)
	}

	// Create list view with options
	w.listView = list.New(
		s.GetEntries(),
		itemsPerPage,
		list.WithShowTags(moduleConfig.EnableTags),
		list.WithHighlight(moduleConfig.EnableHighlight),
		list.WithFavoritesManager(s.GetFavoritesManager()),
	)
	w.listView.OnActivate(w.onAppActivate)
	w.scrolled.SetChild(w.listView.Widget())
	box.Append(w.scrolled)

	// Create pagination info label (only if pagination is enabled)
	if moduleConfig.EnablePagination {
		w.pageLabel = gtk.NewLabel("")
		w.pageLabel.AddCSSClass("page-info")
		w.pageLabel.SetXAlign(0.5)
		w.updatePageLabel()
		box.Append(w.pageLabel)
	}

	w.window.SetChild(box)

	// Set up keyboard event handling
	keyController := gtk.NewEventControllerKey()
	keyController.SetPropagationPhase(gtk.PhaseCapture)
	keyController.ConnectKeyPressed(w.onKeyPressed)
	w.window.AddController(keyController)

	return w
}

// calculateWindowHeight calculates the appropriate window height based on items per page
func (w *Window) calculateWindowHeight() int {
	height := BaseWindowHeight

	// Determine how many rows to show in the window
	var displayRows int
	if w.moduleConfig.EnablePagination {
		// When pagination is enabled, show exactly itemsPerPage rows
		displayRows = w.moduleConfig.ItemsPerPage
	} else {
		// When pagination is disabled, show MaxVisibleRows with scrollbar
		displayRows = MaxVisibleRows
	}

	// Add height for the rows
	height += displayRows * RowHeight

	// Add height for pagination label if enabled
	if w.moduleConfig.EnablePagination {
		height += PageLabelHeight
	}

	return height
}

// onSearchChangedDebounced handles search input with debouncing
func (w *Window) onSearchChangedDebounced() {
	// Cancel previous timer if it exists
	if w.debounceTimer != 0 {
		glib.SourceRemove(w.debounceTimer)
	}

	// Set new timer
	w.debounceTimer = glib.TimeoutAdd(SearchDebounceMs, func() bool {
		w.onSearchChanged()
		w.debounceTimer = 0
		return false // Return false to stop the timer
	})
}

// onSearchChanged handles search input changes
func (w *Window) onSearchChanged() {
	query := w.searchEntry.Text()
	filtered := w.scanner.Filter(query, entry.AppTypeAll)
	w.listView.Update(filtered, query)
	if w.moduleConfig.EnablePagination {
		w.updatePageLabel()
	}
}

// updatePageLabel updates the pagination label
func (w *Window) updatePageLabel() {
	if w.pageLabel != nil {
		w.pageLabel.SetText(w.listView.GetPageInfo())
	}
}

// scrollToSelected ensures the selected row is visible
func (w *Window) scrollToSelected() {
	selected := w.listView.GetSelectedRow()
	if selected != nil {
		vadj := w.scrolled.VAdjustment()
		allocation := selected.Allocation()
		rowY := float64(allocation.Y())
		rowHeight := float64(allocation.Height())
		pageSize := vadj.PageSize()
		currentValue := vadj.Value()

		if rowY < currentValue {
			vadj.SetValue(rowY)
		}

		if rowY+rowHeight > currentValue+pageSize {
			vadj.SetValue(rowY + rowHeight - pageSize)
		}
	}
}

// onKeyPressed handles keyboard shortcuts
func (w *Window) onKeyPressed(keyval uint, _ uint, state gdk.ModifierType) bool {
	switch keyval {
	case gdk.KEY_Escape:
		w.window.Close()
		return true

	case gdk.KEY_Down:
		w.listView.SelectNext()
		w.scrollToSelected()
		return true
	case gdk.KEY_Up:
		w.listView.SelectPrevious()
		w.scrollToSelected()
		return true

	case gdk.KEY_Page_Down, gdk.KEY_Right:
		if w.moduleConfig.EnablePagination {
			w.listView.NextPage()
			w.updatePageLabel()
			w.scrollToSelected()
		}
		return true
	case gdk.KEY_Page_Up, gdk.KEY_Left:
		if w.moduleConfig.EnablePagination {
			w.listView.PreviousPage()
			w.updatePageLabel()
			w.scrollToSelected()
		}
		return true

	case gdk.KEY_Return:
		w.listView.ActivateSelected()
		return true
	}
	return false
}

// onAppActivate handles application launch
func (w *Window) onAppActivate(e *entry.Entry) {
	// Record launch event SYNCHRONOUSLY before closing
	if fm := w.scanner.GetFavoritesManager(); fm != nil {
		fm.RecordLaunch(e)
		// Save immediately (synchronous)
		if err := fm.Save(); err != nil {
			log.Printf("Warning: Failed to save favorites: %v", err)
		}
	}

	// Launch the application
	err := e.Launch()
	if err != nil {
		log.Printf("Error launching %s: %v", e.Name, err)
		return
	}

	// Close window after saving
	w.window.Close()
}

// Shutdown performs cleanup
func (w *Window) Shutdown() {
	// Cleanup is handled by scanner
}

// Show displays the window
func (w *Window) Show() {
	w.window.SetVisible(true)
	w.searchEntry.GrabFocus()
}

// Widget returns the underlying GTK ApplicationWindow (for modules.Window interface)
func (w *Window) Widget() *gtk.ApplicationWindow {
	return w.window
}
