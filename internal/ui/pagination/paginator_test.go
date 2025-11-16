package pagination

import "testing"

func TestNew(t *testing.T) {
	p := New(10)
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.itemsPerPage != 10 {
		t.Errorf("itemsPerPage = %d, want 10", p.itemsPerPage)
	}
}

func TestSetTotalItems(t *testing.T) {
	p := New(10)
	p.SetTotalItems(25)

	if p.totalItems != 25 {
		t.Errorf("totalItems = %d, want 25", p.totalItems)
	}

	if p.TotalPages() != 3 {
		t.Errorf("TotalPages() = %d, want 3", p.TotalPages())
	}
}

func TestNextPage(t *testing.T) {
	p := New(10)
	p.SetTotalItems(25)

	// Should be able to go to page 1
	if !p.NextPage() {
		t.Error("NextPage() should return true")
	}
	if p.CurrentPage() != 1 {
		t.Errorf("CurrentPage() = %d, want 1", p.CurrentPage())
	}

	// Should be able to go to page 2
	if !p.NextPage() {
		t.Error("NextPage() should return true")
	}
	if p.CurrentPage() != 2 {
		t.Errorf("CurrentPage() = %d, want 2", p.CurrentPage())
	}

	// Should not be able to go beyond last page
	if p.NextPage() {
		t.Error("NextPage() should return false at last page")
	}
}

func TestPreviousPage(t *testing.T) {
	p := New(10)
	p.SetTotalItems(25)
	p.NextPage()
	p.NextPage()

	// Should be able to go back
	if !p.PreviousPage() {
		t.Error("PreviousPage() should return true")
	}
	if p.CurrentPage() != 1 {
		t.Errorf("CurrentPage() = %d, want 1", p.CurrentPage())
	}

	p.PreviousPage()

	// Should not be able to go before first page
	if p.PreviousPage() {
		t.Error("PreviousPage() should return false at first page")
	}
}

func TestGetPageItems(t *testing.T) {
	tests := []struct {
		name          string
		itemsPerPage  int
		totalItems    int
		currentPage   int
		expectedStart int
		expectedEnd   int
	}{
		{
			name:          "First page full",
			itemsPerPage:  10,
			totalItems:    25,
			currentPage:   0,
			expectedStart: 0,
			expectedEnd:   10,
		},
		{
			name:          "Second page full",
			itemsPerPage:  10,
			totalItems:    25,
			currentPage:   1,
			expectedStart: 10,
			expectedEnd:   20,
		},
		{
			name:          "Last page partial",
			itemsPerPage:  10,
			totalItems:    25,
			currentPage:   2,
			expectedStart: 20,
			expectedEnd:   25,
		},
		{
			name:          "Single item page",
			itemsPerPage:  10,
			totalItems:    1,
			currentPage:   0,
			expectedStart: 0,
			expectedEnd:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.itemsPerPage)
			p.SetTotalItems(tt.totalItems)
			p.currentPage = tt.currentPage

			start, end := p.GetPageItems()
			if start != tt.expectedStart {
				t.Errorf("start = %d, want %d", start, tt.expectedStart)
			}
			if end != tt.expectedEnd {
				t.Errorf("end = %d, want %d", end, tt.expectedEnd)
			}
		})
	}
}

func TestGetPageInfo(t *testing.T) {
	tests := []struct {
		name         string
		itemsPerPage int
		totalItems   int
		currentPage  int
		expected     string
	}{
		{
			name:         "First page",
			itemsPerPage: 10,
			totalItems:   25,
			currentPage:  0,
			expected:     "Page 1 of 3",
		},
		{
			name:         "Last page",
			itemsPerPage: 10,
			totalItems:   25,
			currentPage:  2,
			expected:     "Page 3 of 3",
		},
		{
			name:         "No items",
			itemsPerPage: 10,
			totalItems:   0,
			currentPage:  0,
			expected:     "No items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.itemsPerPage)
			p.SetTotalItems(tt.totalItems)
			p.currentPage = tt.currentPage

			result := p.GetPageInfo()
			if result != tt.expected {
				t.Errorf("GetPageInfo() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestReset(t *testing.T) {
	p := New(10)
	p.SetTotalItems(25)
	p.NextPage()
	p.NextPage()

	if p.CurrentPage() != 2 {
		t.Fatalf("Setup failed: CurrentPage() = %d, want 2", p.CurrentPage())
	}

	p.Reset()

	if p.CurrentPage() != 0 {
		t.Errorf("After Reset(), CurrentPage() = %d, want 0", p.CurrentPage())
	}
}

func TestTotalPages(t *testing.T) {
	tests := []struct {
		name         string
		itemsPerPage int
		totalItems   int
		expected     int
	}{
		{"Exact division", 10, 30, 3},
		{"Partial page", 10, 25, 3},
		{"Single page", 10, 5, 1},
		{"Empty", 10, 0, 1},
		{"One item", 10, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.itemsPerPage)
			p.SetTotalItems(tt.totalItems)

			result := p.TotalPages()
			if result != tt.expected {
				t.Errorf("TotalPages() = %d, want %d", result, tt.expected)
			}
		})
	}
}
