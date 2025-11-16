package list

import (
	"strings"
	"testing"

	"github.com/antoniosarro/gofi/internal/domain/entry"
)

func TestHighlightText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		query    string
		contains []string // Substrings that should be in result
	}{
		{
			name:     "Simple match",
			text:     "Firefox Browser",
			query:    "fire",
			contains: []string{"<span", "fire", "</span>"},
		},
		{
			name:     "No match",
			text:     "Chrome Browser",
			query:    "fire",
			contains: []string{"Chrome Browser"},
		},
		{
			name:     "Empty query",
			text:     "Firefox Browser",
			query:    "",
			contains: []string{"Firefox Browser"},
		},
		{
			name:     "Case insensitive",
			text:     "Firefox Browser",
			query:    "FIRE",
			contains: []string{"<span", "Fire", "</span>"},
		},
		{
			name:     "Multiple matches",
			text:     "fire fire fire",
			query:    "fire",
			contains: []string{"<span", "fire", "</span>"},
		},
		{
			name:     "Special characters escaped",
			text:     "Test & <HTML>",
			query:    "test",
			contains: []string{"&amp;", "&lt;", "&gt;"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := highlightText(tt.text, tt.query)

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("Result should contain %q, got: %s", substr, result)
				}
			}
		})
	}
}

func TestCreateRow(t *testing.T) {
	e := &entry.Entry{
		Name:    "Firefox",
		Comment: "Web Browser",
		Icon:    "firefox",
		Path:    "/test/firefox.desktop",
	}

	opts := RowOptions{
		ShowTags:        true,
		EnableHighlight: false,
		Query:           "",
	}

	row := createRow(e, opts)

	if row == nil {
		t.Fatal("createRow() returned nil")
	}

	// Basic check that a box was created
	// (Full GTK widget testing would require a display server)
}

func TestCreateRowWithHighlight(t *testing.T) {
	e := &entry.Entry{
		Name:    "Firefox Browser",
		Comment: "Browse the web",
		Icon:    "firefox",
		Path:    "/test/firefox.desktop",
	}

	opts := RowOptions{
		ShowTags:        false,
		EnableHighlight: true,
		Query:           "fire",
	}

	row := createRow(e, opts)

	if row == nil {
		t.Fatal("createRow() returned nil")
	}
}

func TestCreateRowWithTag(t *testing.T) {
	e := &entry.Entry{
		Name: "Firefox",
		Path: "/var/lib/flatpak/exports/share/applications/firefox.desktop",
	}

	opts := RowOptions{
		ShowTags:        true,
		EnableHighlight: false,
		Query:           "",
	}

	row := createRow(e, opts)

	if row == nil {
		t.Fatal("createRow() returned nil")
	}
}
