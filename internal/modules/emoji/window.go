package emoji

import (
	"fmt"
	"log"

	"github.com/antoniosarro/gofi/internal/config"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

const (
	WindowWidth      = 600
	WindowHeight     = 600
	SearchDebounceMs = 150
	RowHeight        = 60
	MaxVisibleRows   = 8
)

type Window struct {
	window        *gtk.ApplicationWindow
	searchEntry   *gtk.SearchEntry
	scrolled      *gtk.ScrolledWindow
	listBox       *gtk.ListBox
	statusLabel   *gtk.Label
	config        *config.ModuleConfig
	emojis        []Emoji
	filtered      []Emoji
	debounceTimer glib.SourceHandle
}

func NewWindow(app *gtk.Application, cfg *config.ModuleConfig, emojis []Emoji) *Window {
	w := &Window{
		window:   gtk.NewApplicationWindow(app),
		config:   cfg,
		emojis:   emojis,
		filtered: emojis,
	}

	w.window.SetTitle("Emoji Picker")
	w.window.SetDefaultSize(WindowWidth, WindowHeight)
	w.window.SetSizeRequest(WindowWidth, WindowHeight)
	w.window.SetDecorated(false)
	w.window.SetResizable(false)

	w.buildUI()
	w.setupKeyBindings()
	w.populateEmojis()

	return w
}

func (w *Window) buildUI() {
	mainBox := gtk.NewBox(gtk.OrientationVertical, 10)
	mainBox.SetMarginTop(10)
	mainBox.SetMarginBottom(10)
	mainBox.SetMarginStart(10)
	mainBox.SetMarginEnd(10)
	mainBox.SetSizeRequest(WindowWidth, WindowHeight)

	gtk.NewSignalListItemFactory()

	// Search entry
	w.searchEntry = gtk.NewSearchEntry()
	w.searchEntry.SetPlaceholderText("Search emojis by name or keyword...")
	w.searchEntry.SetHExpand(true)
	w.searchEntry.ConnectSearchChanged(w.onSearchChangedDebounced)
	mainBox.Append(w.searchEntry)

	// Scrolled window with emoji list
	w.scrolled = gtk.NewScrolledWindow()
	w.scrolled.SetVExpand(true)
	w.scrolled.SetHExpand(true)
	w.scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

	// Set fixed height for scrolled window to fit exactly MaxVisibleRows
	// Account for margins and padding
	scrollHeight := MaxVisibleRows * RowHeight
	w.scrolled.SetSizeRequest(WindowWidth-20, scrollHeight)
	w.scrolled.SetVAlign(gtk.AlignFill)

	w.listBox = gtk.NewListBox()
	w.listBox.SetSelectionMode(gtk.SelectionSingle)
	w.listBox.SetHExpand(true)
	w.listBox.ConnectRowActivated(w.onEmojiActivated)

	w.scrolled.SetChild(w.listBox)
	mainBox.Append(w.scrolled)

	// Status bar
	w.statusLabel = gtk.NewLabel("")
	w.statusLabel.SetXAlign(0)
	w.statusLabel.SetHExpand(true)
	w.statusLabel.AddCSSClass("dim-label")
	mainBox.Append(w.statusLabel)

	w.window.SetChild(mainBox)
}

func (w *Window) setupKeyBindings() {
	keyController := gtk.NewEventControllerKey()
	keyController.SetPropagationPhase(gtk.PhaseCapture)
	keyController.ConnectKeyPressed(w.onKeyPressed)
	w.window.AddController(keyController)
}

func (w *Window) onKeyPressed(keyval uint, _ uint, state gdk.ModifierType) bool {
	switch keyval {
	case gdk.KEY_Escape:
		w.window.Close()
		return true
	case gdk.KEY_Down, gdk.KEY_j:
		w.selectNext()
		return true
	case gdk.KEY_Up, gdk.KEY_k:
		w.selectPrevious()
		return true
	case gdk.KEY_Return:
		w.activateSelected()
		return true
	}
	return false
}

func (w *Window) selectNext() {
	selected := w.listBox.SelectedRow()
	if selected == nil {
		w.listBox.SelectRow(w.listBox.RowAtIndex(0))
		w.scrollToSelected()
		return
	}

	nextIndex := selected.Index() + 1
	if nextIndex < len(w.filtered) {
		w.listBox.SelectRow(w.listBox.RowAtIndex(nextIndex))
		w.scrollToSelected()
	}
}

func (w *Window) selectPrevious() {
	selected := w.listBox.SelectedRow()
	if selected == nil {
		return
	}

	prevIndex := selected.Index() - 1
	if prevIndex >= 0 {
		w.listBox.SelectRow(w.listBox.RowAtIndex(prevIndex))
		w.scrollToSelected()
	}
}

func (w *Window) scrollToSelected() {
	selected := w.listBox.SelectedRow()
	if selected != nil {
		vadj := w.scrolled.VAdjustment()
		allocation := selected.Allocation()

		rowY := float64(allocation.Y())
		rowHeight := float64(allocation.Height())
		pageSize := vadj.PageSize()
		currentValue := vadj.Value()

		// Scroll up if row is above visible area
		if rowY < currentValue {
			vadj.SetValue(rowY)
		}

		// Scroll down if row is below visible area
		if rowY+rowHeight > currentValue+pageSize {
			vadj.SetValue(rowY + rowHeight - pageSize)
		}
	}
}

func (w *Window) onSearchChangedDebounced() {
	if w.debounceTimer != 0 {
		glib.SourceRemove(w.debounceTimer)
	}

	w.debounceTimer = glib.TimeoutAdd(SearchDebounceMs, func() bool {
		w.updateEmojis()
		w.debounceTimer = 0
		return false
	})
}

func (w *Window) updateEmojis() {
	query := w.searchEntry.Text()

	// Search
	w.filtered = Search(w.emojis, query)

	// Update UI
	w.populateEmojis()
	w.updateStatus()
}

func (w *Window) populateEmojis() {
	// Clear existing items
	for {
		child := w.listBox.FirstChild()
		if child == nil {
			break
		}
		w.listBox.Remove(child)
	}

	// Add emoji items
	for i := range w.filtered {
		emoji := &w.filtered[i]
		w.listBox.Append(w.createEmojiRow(emoji))
	}

	// Select first item
	if len(w.filtered) > 0 {
		w.listBox.SelectRow(w.listBox.RowAtIndex(0))
	}
}

func (w *Window) createEmojiRow(emoji *Emoji) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationHorizontal, 12)
	box.SetMarginTop(8)
	box.SetMarginBottom(8)
	box.SetMarginStart(12)
	box.SetMarginEnd(12)
	box.SetHExpand(true)

	// Set fixed height to ensure consistent row height
	box.SetSizeRequest(WindowWidth-40, RowHeight)

	// Emoji character (large)
	emojiLabel := gtk.NewLabel(emoji.Char)
	emojiLabel.SetMarkup("<span size='xx-large'>" + emoji.Char + "</span>")
	emojiLabel.SetSizeRequest(50, -1)
	emojiLabel.SetVAlign(gtk.AlignCenter)
	box.Append(emojiLabel)

	// Text box with name and keywords
	textBox := gtk.NewBox(gtk.OrientationVertical, 4)
	textBox.SetHExpand(true)
	textBox.SetVAlign(gtk.AlignCenter)

	// Emoji name
	nameLabel := gtk.NewLabel(emoji.Name)
	nameLabel.SetXAlign(0)
	nameLabel.SetHExpand(true)
	nameLabel.AddCSSClass("app-name")
	textBox.Append(nameLabel)

	// Keywords/description
	if len(emoji.Keywords) > 0 {
		// Show first few keywords
		keywordsText := ""
		maxKeywords := 5
		for i, kw := range emoji.Keywords {
			if i >= maxKeywords {
				keywordsText += "..."
				break
			}
			if i > 0 {
				keywordsText += ", "
			}
			keywordsText += kw
		}

		keywordsLabel := gtk.NewLabel(keywordsText)
		keywordsLabel.SetXAlign(0)
		keywordsLabel.SetHExpand(true)
		keywordsLabel.SetEllipsize(pango.EllipsizeEnd)
		keywordsLabel.SetMaxWidthChars(60)
		keywordsLabel.AddCSSClass("app-description")
		textBox.Append(keywordsLabel)
	}

	box.Append(textBox)

	// Category tag (optional, on the right)
	if emoji.Category != "" {
		categoryLabel := gtk.NewLabel(emoji.Category)
		categoryLabel.AddCSSClass("app-tag")
		categoryLabel.AddCSSClass("app-tag-System")
		categoryLabel.SetVAlign(gtk.AlignCenter)
		box.Append(categoryLabel)
	}

	return box
}

func (w *Window) onEmojiActivated(row *gtk.ListBoxRow) {
	index := row.Index()
	if index >= 0 && index < len(w.filtered) {
		emoji := w.filtered[index]
		w.copyEmoji(emoji)
	}
}

func (w *Window) activateSelected() {
	selected := w.listBox.SelectedRow()
	if selected != nil {
		w.onEmojiActivated(selected)
	}
}

func (w *Window) copyEmoji(emoji Emoji) {
	err := CopyToClipboard(emoji.Char)
	if err != nil {
		log.Printf("Error copying emoji to clipboard: %v", err)
		w.showError("Failed to copy emoji: " + err.Error())
		return
	}

	log.Printf("Copied emoji to clipboard: %s (%s)", emoji.Char, emoji.Name)
	w.window.Close()
}

func (w *Window) updateStatus() {
	total := len(w.filtered)
	w.statusLabel.SetText(formatStatus(total))
}

func formatStatus(count int) string {
	if count == 0 {
		return "No emojis found"
	}
	if count == 1 {
		return "1 emoji"
	}
	return fmt.Sprintf("%d emojis", count)
}

func (w *Window) showError(message string) {
	dialog := gtk.NewMessageDialog(
		&w.window.Window,
		gtk.DialogModal,
		gtk.MessageError,
		gtk.ButtonsClose,
	)
	dialog.SetMarkup(message)
	dialog.ConnectResponse(func(responseId int) {
		dialog.Close()
	})
	dialog.Show()
}

func (w *Window) Show() {
	w.window.SetVisible(true)
	w.searchEntry.GrabFocus()
}

func (w *Window) Shutdown() {
	if w.debounceTimer != 0 {
		glib.SourceRemove(w.debounceTimer)
		w.debounceTimer = 0
	}
}

func (w *Window) Widget() *gtk.ApplicationWindow {
	return w.window
}
