package list

import (
	"github.com/antoniosarro/gofi/internal/domain/entry"
	"github.com/antoniosarro/gofi/internal/favorites"
	"github.com/antoniosarro/gofi/internal/ui/pagination"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// View handles the application list display with pagination
type View struct {
	listBox          *gtk.ListBox
	entries          []*entry.Entry
	paginator        *pagination.Paginator
	onActivate       func(*entry.Entry)
	showTags         bool
	enableHighlight  bool
	favoritesManager *favorites.Manager
	currentQuery     string
}

// Option is a functional option for View
type Option func(*View)

// WithShowTags enables/disables app type tags
func WithShowTags(show bool) Option {
	return func(v *View) {
		v.showTags = show
	}
}

// WithHighlight enables/disables text highlighting
func WithHighlight(enable bool) Option {
	return func(v *View) {
		v.enableHighlight = enable
	}
}

// WithFavoritesManager sets the favorites manager
func WithFavoritesManager(fm *favorites.Manager) Option {
	return func(v *View) {
		v.favoritesManager = fm
	}
}

// New creates a new list view with pagination
func New(entries []*entry.Entry, itemsPerPage int, opts ...Option) *View {
	v := &View{
		listBox:   gtk.NewListBox(),
		entries:   entries,
		paginator: pagination.New(itemsPerPage),
	}

	// Apply options
	for _, opt := range opts {
		opt(v)
	}

	v.listBox.SetSelectionMode(gtk.SelectionSingle)
	v.listBox.ConnectRowActivated(v.onRowActivated)

	v.paginator.SetTotalItems(len(entries))
	v.populate()

	return v
}

// SetQuery sets the current search query for highlighting
func (v *View) SetQuery(query string) {
	v.currentQuery = query
}

// populate fills the list with entries for the current page
func (v *View) populate() {
	// Clear existing rows
	for {
		row := v.listBox.FirstChild()
		if row == nil {
			break
		}
		v.listBox.Remove(row)
	}

	// Get current page items
	start, end := v.paginator.GetPageItems()

	// Add new rows for current page
	for i := start; i < end && i < len(v.entries); i++ {
		row := createRow(v.entries[i], RowOptions{
			ShowTags:         v.showTags,
			EnableHighlight:  v.enableHighlight,
			Query:            v.currentQuery,
			FavoritesManager: v.favoritesManager,
		})
		v.listBox.Append(row)
	}

	// Select first item if available
	if end > start {
		v.listBox.SelectRow(v.listBox.RowAtIndex(0))
	}
}

// onRowActivated handles row activation
func (v *View) onRowActivated(row *gtk.ListBoxRow) {
	if v.onActivate == nil {
		return
	}

	index := row.Index()
	start, _ := v.paginator.GetPageItems()
	actualIndex := start + index

	if actualIndex >= 0 && actualIndex < len(v.entries) {
		v.onActivate(v.entries[actualIndex])
	}
}

// Update refreshes the list with new entries
func (v *View) Update(entries []*entry.Entry, query string) {
	v.entries = entries
	v.currentQuery = query
	v.paginator.SetTotalItems(len(entries))
	v.paginator.Reset()
	v.populate()
}

// SelectNext selects the next item
func (v *View) SelectNext() {
	selected := v.listBox.SelectedRow()
	if selected == nil {
		v.listBox.SelectRow(v.listBox.RowAtIndex(0))
		return
	}

	nextIndex := selected.Index() + 1
	_, end := v.paginator.GetPageItems()
	start, _ := v.paginator.GetPageItems()
	itemsOnPage := end - start

	if nextIndex < itemsOnPage {
		v.listBox.SelectRow(v.listBox.RowAtIndex(nextIndex))
	}
}

// SelectPrevious selects the previous item
func (v *View) SelectPrevious() {
	selected := v.listBox.SelectedRow()
	if selected == nil {
		return
	}

	prevIndex := selected.Index() - 1
	if prevIndex >= 0 {
		v.listBox.SelectRow(v.listBox.RowAtIndex(prevIndex))
	}
}

// NextPage moves to the next page
func (v *View) NextPage() bool {
	if v.paginator.NextPage() {
		v.populate()
		return true
	}
	return false
}

// PreviousPage moves to the previous page
func (v *View) PreviousPage() bool {
	if v.paginator.PreviousPage() {
		v.populate()
		return true
	}
	return false
}

// GetPageInfo returns pagination info
func (v *View) GetPageInfo() string {
	return v.paginator.GetPageInfo()
}

// GetSelectedRow returns the currently selected row
func (v *View) GetSelectedRow() *gtk.ListBoxRow {
	return v.listBox.SelectedRow()
}

// ActivateSelected activates the currently selected item
func (v *View) ActivateSelected() {
	selected := v.listBox.SelectedRow()
	if selected != nil {
		v.onRowActivated(selected)
	}
}

// OnActivate sets the activation callback
func (v *View) OnActivate(fn func(*entry.Entry)) {
	v.onActivate = fn
}

// Widget returns the underlying GTK widget
func (v *View) Widget() *gtk.Widget {
	return &v.listBox.Widget
}
