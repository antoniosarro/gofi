package pagination

import (
	"fmt"
	"math"
)

// Paginator handles page-based navigation
type Paginator struct {
	itemsPerPage int
	currentPage  int
	totalItems   int
}

// New creates a new paginator
func New(itemsPerPage int) *Paginator {
	return &Paginator{
		itemsPerPage: itemsPerPage,
		currentPage:  0,
		totalItems:   0,
	}
}

// SetTotalItems updates the total number of items
func (p *Paginator) SetTotalItems(total int) {
	p.totalItems = total
	// Reset to first page if current page is out of bounds
	if p.currentPage >= p.TotalPages() {
		p.currentPage = 0
	}
}

// TotalPages returns the total number of pages
func (p *Paginator) TotalPages() int {
	if p.totalItems == 0 {
		return 1
	}
	return int(math.Ceil(float64(p.totalItems) / float64(p.itemsPerPage)))
}

// CurrentPage returns the current page (0-indexed)
func (p *Paginator) CurrentPage() int {
	return p.currentPage
}

// NextPage moves to the next page
func (p *Paginator) NextPage() bool {
	if p.currentPage < p.TotalPages()-1 {
		p.currentPage++
		return true
	}
	return false
}

// PreviousPage moves to the previous page
func (p *Paginator) PreviousPage() bool {
	if p.currentPage > 0 {
		p.currentPage--
		return true
	}
	return false
}

// GetPageItems returns the start and end indices for current page
func (p *Paginator) GetPageItems() (start, end int) {
	start = p.currentPage * p.itemsPerPage
	end = start + p.itemsPerPage
	if end > p.totalItems {
		end = p.totalItems
	}
	return start, end
}

// GetPageInfo returns a formatted page info string
func (p *Paginator) GetPageInfo() string {
	if p.totalItems == 0 {
		return "No items"
	}
	return fmt.Sprintf("Page %d of %d", p.currentPage+1, p.TotalPages())
}

// Reset resets the paginator to the first page
func (p *Paginator) Reset() {
	p.currentPage = 0
}
