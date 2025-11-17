package powermenu

import (
	"log"

	"github.com/antoniosarro/gofi/internal/config"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const (
	WindowWidth    = 400   // Match application launcher width
	WindowHeight   = 350   // Adjusted for 5 rows + header
	RowHeight      = 40    // Match application launcher row height
	UpdateInterval = 60000 // Update uptime every 60 seconds
)

type Window struct {
	window      *gtk.ApplicationWindow
	config      *config.ModuleConfig
	powerMenu   *PowerMenu
	listBox     *gtk.ListBox
	uptimeLabel *gtk.Label
	uptimeTimer glib.SourceHandle
}

func NewWindow(app *gtk.Application, cfg *config.ModuleConfig) *Window {
	w := &Window{
		window:    gtk.NewApplicationWindow(app),
		config:    cfg,
		powerMenu: NewPowerMenu(cfg.Settings),
	}

	w.window.SetTitle("Power Menu")
	w.window.SetDefaultSize(WindowWidth, WindowHeight)
	w.window.SetDecorated(false)
	w.window.SetResizable(false)

	w.buildUI()
	w.setupKeyBindings()
	w.startUptimeTimer()

	return w
}

func (w *Window) buildUI() {
	mainBox := gtk.NewBox(gtk.OrientationVertical, 10)
	mainBox.SetMarginTop(10)
	mainBox.SetMarginBottom(10)
	mainBox.SetMarginStart(10)
	mainBox.SetMarginEnd(10)

	// Header with Power Options (left) and Uptime (right)
	headerBox := gtk.NewBox(gtk.OrientationHorizontal, 10)

	// Left side - Power Options title
	leftBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	leftBox.SetHExpand(true)
	leftBox.SetHAlign(gtk.AlignStart)

	powerIcon := gtk.NewImage()
	powerIcon.SetFromIconName("system-shutdown")
	powerIcon.SetPixelSize(20)
	leftBox.Append(powerIcon)

	titleLabel := gtk.NewLabel("")
	titleLabel.SetMarkup("<span weight='bold'>Power Options</span>")
	titleLabel.SetXAlign(0)
	leftBox.Append(titleLabel)

	headerBox.Append(leftBox)

	// Right side - Uptime
	rightBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	rightBox.SetHAlign(gtk.AlignEnd)

	uptimeBox := gtk.NewBox(gtk.OrientationHorizontal, 2)
	uptimeBox.SetHAlign(gtk.AlignEnd)

	uptimeTitle := gtk.NewLabel("Uptime")
	uptimeTitle.SetXAlign(1)
	uptimeTitle.SetMarkup("<span size='small' weight='bold'>Uptime: </span>")
	uptimeBox.Append(uptimeTitle)

	w.uptimeLabel = gtk.NewLabel(GetSystemUptime())
	w.uptimeLabel.SetXAlign(1)
	w.uptimeLabel.SetMarkup("<span size='small' foreground='#a6adc8'>" + GetSystemUptime() + "</span>")
	uptimeBox.Append(w.uptimeLabel)

	rightBox.Append(uptimeBox)

	uptimeIcon := gtk.NewImage()
	uptimeIcon.SetFromIconName("preferences-system-time")
	uptimeIcon.SetPixelSize(24)
	rightBox.Append(uptimeIcon)

	headerBox.Append(rightBox)

	mainBox.Append(headerBox)

	// Separator
	separator := gtk.NewSeparator(gtk.OrientationHorizontal)
	mainBox.Append(separator)

	// Create list box (no scrolling, fixed height)
	w.listBox = gtk.NewListBox()
	w.listBox.SetSelectionMode(gtk.SelectionSingle)
	w.listBox.ConnectRowActivated(w.onActionActivated)
	w.listBox.SetVExpand(true)

	// Add power actions
	for _, action := range w.powerMenu.Actions {
		row := w.createActionRow(action)
		w.listBox.Append(row)
	}

	// Select first row by default
	w.listBox.SelectRow(w.listBox.RowAtIndex(0))

	mainBox.Append(w.listBox)

	w.window.SetChild(mainBox)
}

func (w *Window) createActionRow(action PowerAction) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationHorizontal, 12)
	box.SetMarginTop(8)
	box.SetMarginBottom(8)
	box.SetMarginStart(10)
	box.SetMarginEnd(10)

	// Set minimum height to match application launcher rows
	box.SetSizeRequest(-1, RowHeight)

	// Icon
	icon := gtk.NewImage()
	icon.SetFromIconName(action.Icon)
	icon.SetPixelSize(30)
	box.Append(icon)

	// Text content
	textBox := gtk.NewBox(gtk.OrientationVertical, 4)
	textBox.SetHExpand(true)
	textBox.SetVAlign(gtk.AlignCenter)

	// Action name
	nameLabel := gtk.NewLabel(action.Name)
	nameLabel.SetXAlign(0)
	nameLabel.SetMarkup("<span size='24px' weight='bold'>" + action.Name + "</span>")
	nameLabel.AddCSSClass("app-name")
	textBox.Append(nameLabel)

	// Command (shown in smaller text)
	if action.Command != "" {
		cmdLabel := gtk.NewLabel(action.Command)
		cmdLabel.SetXAlign(0)
		cmdLabel.SetMarkup("<span size='20px' foreground='#a6adc8'>" + action.Command + "</span>")
		cmdLabel.SetEllipsize(3)
		cmdLabel.SetMaxWidthChars(60)
		cmdLabel.AddCSSClass("app-description")
		textBox.Append(cmdLabel)
	}

	box.Append(textBox)

	return box
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
	case gdk.KEY_Return, gdk.KEY_space:
		w.activateSelected()
		return true
	}
	return false
}

func (w *Window) selectNext() {
	selected := w.listBox.SelectedRow()
	if selected == nil {
		w.listBox.SelectRow(w.listBox.RowAtIndex(0))
		return
	}

	nextIndex := selected.Index() + 1
	if nextIndex < len(w.powerMenu.Actions) {
		w.listBox.SelectRow(w.listBox.RowAtIndex(nextIndex))
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
	}
}

func (w *Window) activateSelected() {
	selected := w.listBox.SelectedRow()
	if selected != nil {
		w.onActionActivated(selected)
	}
}

func (w *Window) onActionActivated(row *gtk.ListBoxRow) {
	index := row.Index()
	if index < 0 || index >= len(w.powerMenu.Actions) {
		return
	}

	action := w.powerMenu.Actions[index]

	// Show confirmation dialog for destructive actions
	if action.Name == "Restart" || action.Name == "Shutdown" {
		w.showConfirmationDialog(action, func(confirmed bool) {
			if confirmed {
				// w.executeAction(action)
			}
		})
		return
	}

	w.executeAction(action)
}

func (w *Window) executeAction(action PowerAction) {
	log.Printf("Executing power action: %s (%s)", action.Name, action.Command)

	err := w.powerMenu.ExecuteAction(action.Name)
	if err != nil {
		log.Printf("Error executing %s: %v", action.Name, err)
		w.showErrorDialog(action.Name, err.Error())
		return
	}

	w.window.Close()
}

func (w *Window) showConfirmationDialog(action PowerAction, callback func(bool)) {
	dialog := gtk.NewWindow()
	dialog.SetTransientFor(&w.window.Window)
	dialog.SetModal(true)
	dialog.SetTitle("Confirm " + action.Name)
	dialog.SetDefaultSize(300, 200)
	dialog.SetResizable(false)

	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginTop(20)
	box.SetMarginBottom(20)
	box.SetMarginStart(20)
	box.SetMarginEnd(20)

	// Icon
	iconBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	iconBox.SetHAlign(gtk.AlignCenter)
	icon := gtk.NewImage()
	icon.SetFromIconName("dialog-warning")
	icon.SetPixelSize(40)
	iconBox.Append(icon)
	box.Append(iconBox)

	// Message
	messageLabel := gtk.NewLabel("Are you sure you want to " + action.Name + "?")
	messageLabel.SetHAlign(gtk.AlignCenter)
	messageLabel.SetWrap(true)
	box.Append(messageLabel)

	// Command
	cmdLabel := gtk.NewLabel("")
	cmdLabel.SetMarkup("<span size='small' foreground='#a6adc8'>Command: " + action.Command + "</span>")
	cmdLabel.SetHAlign(gtk.AlignCenter)
	cmdLabel.SetWrap(true)
	box.Append(cmdLabel)

	// Buttons
	buttonBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	buttonBox.SetHAlign(gtk.AlignCenter)
	buttonBox.SetMarginTop(10)

	cancelButton := gtk.NewButtonWithLabel("Cancel")
	cancelButton.SetSizeRequest(100, -1)
	cancelButton.ConnectClicked(func() {
		dialog.Close()
		callback(false)
	})
	buttonBox.Append(cancelButton)

	confirmButton := gtk.NewButtonWithLabel(action.Name)
	confirmButton.SetSizeRequest(100, -1)
	confirmButton.AddCSSClass("destructive-action")
	confirmButton.ConnectClicked(func() {
		dialog.Close()
		callback(true)
	})
	buttonBox.Append(confirmButton)

	box.Append(buttonBox)
	dialog.SetChild(box)

	// Handle Escape key
	keyController := gtk.NewEventControllerKey()
	keyController.ConnectKeyPressed(func(keyval uint, _ uint, _ gdk.ModifierType) bool {
		if keyval == gdk.KEY_Escape {
			dialog.Close()
			callback(false)
			return true
		}
		return false
	})
	dialog.AddController(keyController)

	dialog.Present()
}

func (w *Window) showErrorDialog(actionName string, errorMsg string) {
	dialog := gtk.NewWindow()
	dialog.SetTransientFor(&w.window.Window)
	dialog.SetModal(true)
	dialog.SetTitle("Error")
	dialog.SetDefaultSize(400, 200)
	dialog.SetResizable(false)

	box := gtk.NewBox(gtk.OrientationVertical, 20)
	box.SetMarginTop(20)
	box.SetMarginBottom(20)
	box.SetMarginStart(20)
	box.SetMarginEnd(20)

	// Icon
	iconBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	iconBox.SetHAlign(gtk.AlignCenter)
	icon := gtk.NewImage()
	icon.SetFromIconName("dialog-error")
	icon.SetPixelSize(48)
	iconBox.Append(icon)
	box.Append(iconBox)

	// Title
	titleLabel := gtk.NewLabel("")
	titleLabel.SetMarkup("<span size='large' weight='bold'>Error: " + actionName + "</span>")
	titleLabel.SetHAlign(gtk.AlignCenter)
	box.Append(titleLabel)

	// Error message
	errorLabel := gtk.NewLabel("Failed to execute " + actionName)
	errorLabel.SetHAlign(gtk.AlignCenter)
	box.Append(errorLabel)

	// Detailed error
	detailLabel := gtk.NewLabel(errorMsg)
	detailLabel.SetHAlign(gtk.AlignCenter)
	detailLabel.SetWrap(true)
	detailLabel.SetMaxWidthChars(50)
	detailLabel.AddCSSClass("dim-label")
	box.Append(detailLabel)

	// OK Button
	buttonBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	buttonBox.SetHAlign(gtk.AlignCenter)
	buttonBox.SetMarginTop(10)

	okButton := gtk.NewButtonWithLabel("OK")
	okButton.SetSizeRequest(100, -1)
	okButton.ConnectClicked(func() {
		dialog.Close()
	})
	buttonBox.Append(okButton)

	box.Append(buttonBox)
	dialog.SetChild(box)

	// Handle Escape key
	keyController := gtk.NewEventControllerKey()
	keyController.ConnectKeyPressed(func(keyval uint, _ uint, _ gdk.ModifierType) bool {
		if keyval == gdk.KEY_Escape {
			dialog.Close()
			return true
		}
		return false
	})
	dialog.AddController(keyController)

	dialog.Present()
}

func (w *Window) startUptimeTimer() {
	w.uptimeTimer = glib.TimeoutAdd(UpdateInterval, func() bool {
		w.updateUptime()
		return true // Continue timer
	})
}

func (w *Window) updateUptime() {
	uptime := GetSystemUptime()
	w.uptimeLabel.SetMarkup("<span size='small' foreground='#a6adc8'>" + uptime + "</span>")
}

func (w *Window) Show() {
	w.window.SetVisible(true)
}

func (w *Window) Shutdown() {
	if w.uptimeTimer != 0 {
		glib.SourceRemove(w.uptimeTimer)
		w.uptimeTimer = 0
	}
}

func (w *Window) Widget() *gtk.ApplicationWindow {
	return w.window
}
