package search

import (
	"strings"
	"testing"

	"github.com/antoniosarro/gofi/internal/domain/entry"
)

func TestNewIndexer(t *testing.T) {
	indexer := NewIndexer()
	if indexer == nil {
		t.Fatal("NewIndexer() returned nil")
	}
	if indexer.Count() != 0 {
		t.Errorf("New indexer should have 0 indices, got %d", indexer.Count())
	}
}

func TestIndexerAdd(t *testing.T) {
	indexer := NewIndexer()
	e := &entry.Entry{
		Name:        "Firefox",
		GenericName: "Web Browser",
		Comment:     "Browse the web",
		Path:        "/usr/share/applications/firefox.desktop",
		Categories:  []string{"Network", "WebBrowser"},
	}

	indexer.Add(e)

	if indexer.Count() != 1 {
		t.Errorf("Count() = %d, want 1", indexer.Count())
	}

	index := indexer.Get(e.Path)
	if index == nil {
		t.Fatal("Get() returned nil")
	}

	if index.Entry != e {
		t.Error("Index entry doesn't match original")
	}

	if index.NameNormalized != "firefox" {
		t.Errorf("NameNormalized = %q, want %q", index.NameNormalized, "firefox")
	}

	if !strings.Contains(index.SearchableText, "Firefox") {
		t.Error("SearchableText should contain name")
	}

	if !strings.Contains(index.SearchableText, "Web Browser") {
		t.Error("SearchableText should contain generic name")
	}
}

func TestIndexerBuild(t *testing.T) {
	entries := []*entry.Entry{
		{Name: "Firefox", Path: "/path/firefox.desktop"},
		{Name: "Chrome", Path: "/path/chrome.desktop"},
		{Name: "Safari", Path: "/path/safari.desktop"},
	}

	indexer := NewIndexer()
	indexer.Build(entries)

	if indexer.Count() != 3 {
		t.Errorf("Count() = %d, want 3", indexer.Count())
	}

	for _, e := range entries {
		if indexer.Get(e.Path) == nil {
			t.Errorf("Entry %s not indexed", e.Name)
		}
	}
}

func TestIndexerRemove(t *testing.T) {
	indexer := NewIndexer()
	e := &entry.Entry{
		Name: "Firefox",
		Path: "/path/firefox.desktop",
	}

	indexer.Add(e)
	if indexer.Count() != 1 {
		t.Fatal("Failed to add entry")
	}

	indexer.Remove(e.Path)
	if indexer.Count() != 0 {
		t.Error("Remove() did not remove entry")
	}

	if indexer.Get(e.Path) != nil {
		t.Error("Get() should return nil after Remove()")
	}
}

func TestIndexerClear(t *testing.T) {
	entries := []*entry.Entry{
		{Name: "Firefox", Path: "/path/firefox.desktop"},
		{Name: "Chrome", Path: "/path/chrome.desktop"},
	}

	indexer := NewIndexer()
	indexer.Build(entries)

	indexer.Clear()
	if indexer.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", indexer.Count())
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "Simple text",
			text:     "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "With punctuation",
			text:     "hello, world!",
			expected: []string{"hello", "world"},
		},
		{
			name:     "Mixed case",
			text:     "Hello World",
			expected: []string{"hello", "world"},
		},
		{
			name:     "With numbers",
			text:     "version 3.14",
			expected: []string{"version", "3", "14"},
		},
		{
			name:     "Empty string",
			text:     "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenize(tt.text)
			if len(result) != len(tt.expected) {
				t.Errorf("tokenize() length = %d, want %d", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("tokenize()[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestSearchableTextLongComment(t *testing.T) {
	longComment := strings.Repeat("word ", 20) // 20 words

	indexer := NewIndexer()
	e := &entry.Entry{
		Name:    "Test",
		Comment: longComment,
		Path:    "/test.desktop",
	}

	indexer.Add(e)
	index := indexer.Get(e.Path)

	// Should only include first 10 words of comment
	commentWords := strings.Fields(longComment)
	expectedWords := commentWords[:10]

	for _, word := range expectedWords {
		if !strings.Contains(index.SearchableText, word) {
			t.Errorf("SearchableText missing word from first 10: %s", word)
		}
	}
}
