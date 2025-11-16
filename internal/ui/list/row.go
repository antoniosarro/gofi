package list

import (
	"strings"

	"github.com/antoniosarro/gofi/internal/domain/entry"
	"github.com/antoniosarro/gofi/internal/favorites"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

// RowOptions contains options for row rendering
type RowOptions struct {
	ShowTags         bool
	EnableHighlight  bool
	Query            string
	FavoritesManager *favorites.Manager
}

// createRow creates a list row for an entry
func createRow(e *entry.Entry, opts RowOptions) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationHorizontal, 12)
	box.SetMarginTop(8)
	box.SetMarginBottom(8)
	box.SetMarginStart(12)
	box.SetMarginEnd(12)

	// Icon
	icon := gtk.NewImage()
	if e.Icon != "" {
		icon.SetFromIconName(e.Icon)
	} else {
		icon.SetFromIconName("application-x-executable")
	}
	icon.SetPixelSize(32)
	box.Append(icon)

	// Text container
	textBox := gtk.NewBox(gtk.OrientationVertical, 4)
	textBox.SetHExpand(true)

	// Name with highlighting
	nameLabel := gtk.NewLabel("")
	nameLabel.SetXAlign(0)
	nameLabel.AddCSSClass("app-name")

	if opts.EnableHighlight && opts.Query != "" {
		nameLabel.SetMarkup(highlightText(e.Name, opts.Query))
	} else {
		nameLabel.SetText(e.Name)
	}

	textBox.Append(nameLabel)

	// Description with highlighting
	if e.Comment != "" {
		descLabel := gtk.NewLabel("")
		descLabel.SetXAlign(0)
		descLabel.SetEllipsize(pango.EllipsizeEnd)
		descLabel.SetMaxWidthChars(60)
		descLabel.AddCSSClass("app-description")

		if opts.EnableHighlight && opts.Query != "" {
			descLabel.SetMarkup(highlightText(e.Comment, opts.Query))
		} else {
			descLabel.SetText(e.Comment)
		}

		textBox.Append(descLabel)
	}

	box.Append(textBox)

	// Favorite star icon
	if opts.FavoritesManager != nil && opts.FavoritesManager.IsFavorite(e) {
		starIcon := gtk.NewImage()
		starIcon.SetFromIconName("starred-symbolic")
		starIcon.SetPixelSize(16)
		starIcon.SetTooltipText("Favorite")
		starIcon.SetVAlign(gtk.AlignCenter)
		box.Append(starIcon)
	}

	// App type tag (only if enabled)
	if opts.ShowTags {
		appType := e.GetAppType()
		if appType != entry.AppTypeOther {
			tag := gtk.NewLabel(string(appType))
			tag.AddCSSClass("app-tag")
			tag.AddCSSClass("app-tag-" + string(appType))
			tag.SetVAlign(gtk.AlignCenter)
			box.Append(tag)
		}
	}

	return box
}

// highlightText highlights the query in the text using Pango markup
func highlightText(text, query string) string {
	if query == "" {
		return text
	}

	// Escape text for Pango markup
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")

	queryLower := strings.ToLower(query)
	textLower := strings.ToLower(text)

	// Find all occurrences of the query
	result := ""
	lastIndex := 0

	for {
		index := strings.Index(textLower[lastIndex:], queryLower)
		if index == -1 {
			// No more matches, append the rest
			result += text[lastIndex:]
			break
		}

		// Adjust index to be relative to the original string
		index += lastIndex

		// Append text before match
		result += text[lastIndex:index]

		// Append highlighted match
		matchedText := text[index : index+len(query)]
		result += `<span background="#89b4fa" foreground="#1e1e2e" weight="bold">` + matchedText + `</span>`

		// Move past this match
		lastIndex = index + len(query)
	}

	return result
}
