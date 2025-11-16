package search

import (
	"testing"

	"github.com/antoniosarro/gofi/internal/domain/entry"
	"github.com/antoniosarro/gofi/internal/search/fuzzy"
)

func TestNew(t *testing.T) {
	entries := []*entry.Entry{
		{Name: "Firefox", Path: "/path/firefox.desktop"},
		{Name: "Chrome", Path: "/path/chrome.desktop"},
	}

	engine := New(entries)

	if engine == nil {
		t.Fatal("New() returned nil")
	}

	if engine.indexer.Count() != 2 {
		t.Errorf("Indexer count = %d, want 2", engine.indexer.Count())
	}
}

func TestNewWithOptions(t *testing.T) {
	entries := []*entry.Entry{
		{Name: "Firefox", Path: "/path/firefox.desktop"},
	}

	customMatcher := fuzzy.New(fuzzy.WithCaseSensitive(true))
	engine := New(entries, WithFuzzyMatcher(customMatcher))

	if engine.fuzzyMatcher != customMatcher {
		t.Error("Custom fuzzy matcher was not set")
	}
}

func TestSearchExactMatch(t *testing.T) {
	entries := []*entry.Entry{
		{Name: "Firefox", Path: "/path/firefox.desktop", Exec: "firefox"},
		{Name: "Thunderbird", Path: "/path/thunderbird.desktop", Exec: "thunderbird"},
	}

	engine := New(entries)
	results := engine.Search("Firefox", entry.AppTypeAll, entries)

	if len(results) == 0 {
		t.Fatal("Search() returned no results")
	}

	if results[0].Name != "Firefox" {
		t.Errorf("First result = %s, want Firefox", results[0].Name)
	}
}

func TestSearchPrefixMatch(t *testing.T) {
	entries := []*entry.Entry{
		{Name: "Firefox", Path: "/path/firefox.desktop", Exec: "firefox"},
		{Name: "FileZilla", Path: "/path/filezilla.desktop", Exec: "filezilla"},
	}

	engine := New(entries)
	results := engine.Search("fire", entry.AppTypeAll, entries)

	if len(results) == 0 {
		t.Fatal("Search() returned no results for prefix match")
	}

	if results[0].Name != "Firefox" {
		t.Errorf("First result = %s, want Firefox", results[0].Name)
	}
}

func TestSearchFuzzyMatch(t *testing.T) {
	entries := []*entry.Entry{
		{Name: "Firefox", Path: "/path/firefox.desktop", Exec: "firefox"},
		{Name: "Chrome", Path: "/path/chrome.desktop", Exec: "chrome"},
	}

	engine := New(entries)
	results := engine.Search("ffx", entry.AppTypeAll, entries)

	if len(results) == 0 {
		t.Fatal("Search() returned no results for fuzzy match")
	}

	// Should match Firefox with fuzzy matching
	found := false
	for _, r := range results {
		if r.Name == "Firefox" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Fuzzy match did not find Firefox")
	}
}

func TestSearchByAppType(t *testing.T) {
	entries := []*entry.Entry{
		{
			Name:       "Firefox",
			Path:       "/var/lib/flatpak/exports/share/applications/firefox.desktop",
			Exec:       "firefox",
			Categories: []string{"Network"},
		},
		{
			Name:       "Alacritty",
			Path:       "/run/current-system/sw/share/applications/alacritty.desktop",
			Exec:       "alacritty",
			Categories: []string{"System"},
		},
		{
			Name:       "SuperTux",
			Path:       "/usr/share/applications/supertux.desktop",
			Exec:       "supertux",
			Categories: []string{"Game"},
		},
	}

	engine := New(entries)

	// Search for games only
	results := engine.Search("", entry.AppTypeGame, entries)

	if len(results) != 1 {
		t.Errorf("Search(AppTypeGame) returned %d results, want 1", len(results))
	}

	if len(results) > 0 && results[0].Name != "SuperTux" {
		t.Errorf("Game search result = %s, want SuperTux", results[0].Name)
	}

	// Search for Flatpak only
	results = engine.Search("", entry.AppTypeFlatpak, entries)

	if len(results) != 1 {
		t.Errorf("Search(AppTypeFlatpak) returned %d results, want 1", len(results))
	}

	if len(results) > 0 && results[0].Name != "Firefox" {
		t.Errorf("Flatpak search result = %s, want Firefox", results[0].Name)
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	entries := []*entry.Entry{
		{Name: "Firefox", Path: "/path/firefox.desktop", Exec: "firefox"},
		{Name: "Chrome", Path: "/path/chrome.desktop", Exec: "chrome"},
	}

	engine := New(entries)
	results := engine.Search("", entry.AppTypeAll, entries)

	if len(results) != 2 {
		t.Errorf("Search(\"\") returned %d results, want 2", len(results))
	}
}

func TestSearchMinimumScore(t *testing.T) {
	entries := []*entry.Entry{
		{Name: "Firefox", Path: "/path/firefox.desktop", Exec: "firefox"},
		{Name: "Completely Different App", Path: "/path/other.desktop", Exec: "other"},
	}

	engine := New(entries)
	// Search with query that shouldn't match the second entry well
	results := engine.Search("fire", entry.AppTypeAll, entries)

	// Should only return Firefox (above minimum score)
	if len(results) != 1 {
		t.Errorf("Search() returned %d results, expected filtering by minimum score", len(results))
	}

	if len(results) > 0 && results[0].Name != "Firefox" {
		t.Errorf("Result = %s, want Firefox", results[0].Name)
	}
}

func TestSearchCommentMatch(t *testing.T) {
	entries := []*entry.Entry{
		{
			Name:    "MyApp",
			Comment: "A powerful web browser",
			Path:    "/path/myapp.desktop",
			Exec:    "myapp",
		},
		{
			Name:    "OtherApp",
			Comment: "A simple text editor",
			Path:    "/path/other.desktop",
			Exec:    "other",
		},
	}

	engine := New(entries)
	results := engine.Search("browser", entry.AppTypeAll, entries)

	if len(results) == 0 {
		t.Fatal("Search() found no results for comment match")
	}

	if results[0].Name != "MyApp" {
		t.Errorf("Comment search result = %s, want MyApp", results[0].Name)
	}
}

func TestSearchGenericNameMatch(t *testing.T) {
	entries := []*entry.Entry{
		{
			Name:        "firefox",
			GenericName: "Web Browser",
			Path:        "/path/firefox.desktop",
			Exec:        "firefox",
		},
		{
			Name: "gedit",
			Path: "/path/gedit.desktop",
			Exec: "gedit",
		},
	}

	engine := New(entries)
	results := engine.Search("browser", entry.AppTypeAll, entries)

	if len(results) == 0 {
		t.Fatal("Search() found no results for generic name match")
	}

	found := false
	for _, r := range results {
		if r.Name == "firefox" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Generic name match did not find firefox")
	}
}

func TestSearchCategoryMatch(t *testing.T) {
	entries := []*entry.Entry{
		{
			Name:       "Firefox",
			Categories: []string{"Network", "WebBrowser"},
			Path:       "/path/firefox.desktop",
			Exec:       "firefox",
		},
		{
			Name:       "Gimp",
			Categories: []string{"Graphics", "RasterGraphics"},
			Path:       "/path/gimp.desktop",
			Exec:       "gimp",
		},
	}

	engine := New(entries)
	results := engine.Search("network", entry.AppTypeAll, entries)

	if len(results) == 0 {
		t.Fatal("Search() found no results for category match")
	}

	found := false
	for _, r := range results {
		if r.Name == "Firefox" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Category match did not find Firefox")
	}
}

func TestUpdateIndex(t *testing.T) {
	entries := []*entry.Entry{
		{Name: "Firefox", Path: "/path/firefox.desktop", Exec: "firefox"},
	}

	engine := New(entries)

	newEntry := &entry.Entry{
		Name: "Chrome",
		Path: "/path/chrome.desktop",
		Exec: "chrome",
	}

	engine.UpdateIndex(newEntry)

	if engine.indexer.Get(newEntry.Path) == nil {
		t.Error("UpdateIndex() did not add new entry to index")
	}
}

func TestRemoveIndex(t *testing.T) {
	entries := []*entry.Entry{
		{Name: "Firefox", Path: "/path/firefox.desktop", Exec: "firefox"},
	}

	engine := New(entries)
	engine.RemoveIndex("/path/firefox.desktop")

	if engine.indexer.Get("/path/firefox.desktop") != nil {
		t.Error("RemoveIndex() did not remove entry from index")
	}
}

func TestMatchTokens(t *testing.T) {
	tests := []struct {
		name         string
		queryTokens  []string
		targetTokens []string
		expected     bool
	}{
		{
			name:         "Match found",
			queryTokens:  []string{"web", "browser"},
			targetTokens: []string{"powerful", "web", "application"},
			expected:     true,
		},
		{
			name:         "No match",
			queryTokens:  []string{"editor"},
			targetTokens: []string{"browser", "internet"},
			expected:     false,
		},
		{
			name:         "Partial match",
			queryTokens:  []string{"brows"},
			targetTokens: []string{"browser"},
			expected:     true,
		},
		{
			name:         "Empty query",
			queryTokens:  []string{},
			targetTokens: []string{"browser"},
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchTokens(tt.queryTokens, tt.targetTokens)
			if result != tt.expected {
				t.Errorf("matchTokens() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFilterByType(t *testing.T) {
	entries := []*entry.Entry{
		{Name: "Firefox", Path: "/var/lib/flatpak/exports/share/applications/firefox.desktop"},
		{Name: "Chrome", Path: "/var/lib/flatpak/exports/share/applications/chrome.desktop"},
		{Name: "Alacritty", Path: "/run/current-system/sw/share/applications/alacritty.desktop"},
	}

	tests := []struct {
		name     string
		appType  entry.AppType
		expected int
	}{
		{"All entries", entry.AppTypeAll, 3},
		{"Flatpak only", entry.AppTypeFlatpak, 2},
		{"NixSystem only", entry.AppTypeNixSystem, 1},
		{"Games only", entry.AppTypeGame, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterByType(entries, tt.appType)
			if len(result) != tt.expected {
				t.Errorf("filterByType(%v) length = %d, want %d", tt.appType, len(result), tt.expected)
			}
		})
	}
}

func BenchmarkSearch(b *testing.B) {
	entries := make([]*entry.Entry, 100)
	for i := 0; i < 100; i++ {
		entries[i] = &entry.Entry{
			Name: "Application " + string(rune('A'+i%26)),
			Path: "/path/app" + string(rune('0'+i)),
			Exec: "app" + string(rune('0'+i)),
		}
	}

	engine := New(entries)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Search("app", entry.AppTypeAll, entries)
	}
}

func BenchmarkSearchFuzzy(b *testing.B) {
	entries := make([]*entry.Entry, 100)
	for i := 0; i < 100; i++ {
		entries[i] = &entry.Entry{
			Name: "Application " + string(rune('A'+i%26)),
			Path: "/path/app" + string(rune('0'+i)),
			Exec: "app" + string(rune('0'+i)),
		}
	}

	engine := New(entries)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Search("apc", entry.AppTypeAll, entries)
	}
}
