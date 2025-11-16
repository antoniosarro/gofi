package search

import (
	"testing"

	"github.com/antoniosarro/gofi/internal/domain/entry"
)

func TestSearchEngineIntegration(t *testing.T) {
	// Create a realistic set of applications
	entries := []*entry.Entry{
		{
			Name:        "Firefox",
			GenericName: "Web Browser",
			Comment:     "Browse the World Wide Web",
			Path:        "/var/lib/flatpak/exports/share/applications/org.mozilla.firefox.desktop",
			Exec:        "firefox",
			Categories:  []string{"Network", "WebBrowser"},
		},
		{
			Name:        "Firefox Developer Edition",
			GenericName: "Web Browser",
			Comment:     "Firefox for developers",
			Path:        "/usr/share/applications/firefox-dev.desktop",
			Exec:        "firefox-developer-edition",
			Categories:  []string{"Network", "WebBrowser", "Development"},
		},
		{
			Name:        "Thunderbird",
			GenericName: "Mail Client",
			Comment:     "Read and write emails",
			Path:        "/usr/share/applications/thunderbird.desktop",
			Exec:        "thunderbird",
			Categories:  []string{"Network", "Email"},
		},
		{
			Name:        "Alacritty",
			GenericName: "Terminal",
			Comment:     "A fast, cross-platform, OpenGL terminal emulator",
			Path:        "/run/current-system/sw/share/applications/alacritty.desktop",
			Exec:        "alacritty",
			Categories:  []string{"System", "TerminalEmulator"},
		},
		{
			Name:        "Visual Studio Code",
			GenericName: "Code Editor",
			Comment:     "Code editing. Redefined.",
			Path:        "/usr/share/applications/code.desktop",
			Exec:        "code",
			Categories:  []string{"Development", "IDE"},
		},
		{
			Name:        "GIMP",
			GenericName: "Image Editor",
			Comment:     "Create images and edit photographs",
			Path:        "/usr/share/applications/gimp.desktop",
			Exec:        "gimp",
			Categories:  []string{"Graphics", "RasterGraphics"},
		},
		{
			Name:        "SuperTuxKart",
			GenericName: "Racing Game",
			Comment:     "A kart racing game",
			Path:        "/usr/share/applications/supertuxkart.desktop",
			Exec:        "supertuxkart",
			Categories:  []string{"Game", "ArcadeGame"},
		},
	}

	engine := New(entries)

	// Test 1: Exact match
	t.Run("Exact match", func(t *testing.T) {
		results := engine.Search("Firefox", entry.AppTypeAll, entries)
		if len(results) == 0 {
			t.Fatal("No results for exact match")
		}
		if results[0].Name != "Firefox" {
			t.Errorf("First result = %s, want Firefox", results[0].Name)
		}
	})

	// Test 2: Prefix match
	t.Run("Prefix match", func(t *testing.T) {
		results := engine.Search("fire", entry.AppTypeAll, entries)
		if len(results) < 2 {
			t.Fatal("Expected at least 2 results for 'fire'")
		}
		// Both Firefox entries should match
		firefoxCount := 0
		for _, r := range results {
			if r.Name == "Firefox" || r.Name == "Firefox Developer Edition" {
				firefoxCount++
			}
		}
		if firefoxCount < 2 {
			t.Errorf("Found %d Firefox entries, want 2", firefoxCount)
		}
	})

	// Test 3: Fuzzy match
	t.Run("Fuzzy match", func(t *testing.T) {
		results := engine.Search("ffx", entry.AppTypeAll, entries)
		if len(results) == 0 {
			t.Fatal("No results for fuzzy match")
		}
		// Should match Firefox
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
	})

	// Test 4: Comment/description match
	t.Run("Comment match", func(t *testing.T) {
		results := engine.Search("terminal emulator", entry.AppTypeAll, entries)
		if len(results) == 0 {
			t.Fatal("No results for comment match")
		}
		// Should match Alacritty
		found := false
		for _, r := range results {
			if r.Name == "Alacritty" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Comment match did not find Alacritty")
		}
	})

	// Test 5: Category match
	t.Run("Category match", func(t *testing.T) {
		results := engine.Search("development", entry.AppTypeAll, entries)
		if len(results) == 0 {
			t.Fatal("No results for category match")
		}
		// Should match VS Code and Firefox Dev Edition
		vscodeFound := false
		for _, r := range results {
			if r.Name == "Visual Studio Code" {
				vscodeFound = true
				break
			}
		}
		if !vscodeFound {
			t.Error("Category match did not find VS Code")
		}
	})

	// Test 6: Filter by app type
	t.Run("Filter by Flatpak", func(t *testing.T) {
		results := engine.Search("", entry.AppTypeFlatpak, entries)
		if len(results) != 1 {
			t.Errorf("Flatpak filter returned %d results, want 1", len(results))
		}
		if len(results) > 0 && results[0].Name != "Firefox" {
			t.Errorf("Flatpak result = %s, want Firefox", results[0].Name)
		}
	})

	// Test 7: Filter by game type
	t.Run("Filter by Games", func(t *testing.T) {
		results := engine.Search("", entry.AppTypeGame, entries)
		if len(results) != 1 {
			t.Errorf("Game filter returned %d results, want 1", len(results))
		}
		if len(results) > 0 && results[0].Name != "SuperTuxKart" {
			t.Errorf("Game result = %s, want SuperTuxKart", results[0].Name)
		}
	})

	// Test 8: Combined query and type filter
	t.Run("Query with type filter", func(t *testing.T) {
		results := engine.Search("browser", entry.AppTypeFlatpak, entries)
		if len(results) != 1 {
			t.Errorf("Combined filter returned %d results, want 1", len(results))
		}
		if len(results) > 0 && results[0].Name != "Firefox" {
			t.Errorf("Combined result = %s, want Firefox", results[0].Name)
		}
	})

	// Test 9: No match scenario
	t.Run("No match", func(t *testing.T) {
		results := engine.Search("nonexistentapp12345", entry.AppTypeAll, entries)
		if len(results) != 0 {
			t.Errorf("No match should return 0 results, got %d", len(results))
		}
	})

	// Test 10: Generic name match
	t.Run("Generic name match", func(t *testing.T) {
		results := engine.Search("terminal", entry.AppTypeAll, entries)
		found := false
		for _, r := range results {
			if r.Name == "Alacritty" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Generic name match did not find Alacritty")
		}
	})
}

func TestRankingOrder(t *testing.T) {
	entries := []*entry.Entry{
		{Name: "Firefox", Path: "/path/firefox.desktop", Exec: "firefox"},
		{Name: "Firefox Developer", Path: "/path/firefox-dev.desktop", Exec: "firefox-dev"},
		{Name: "File Manager", Path: "/path/filemanager.desktop", Exec: "filemanager"},
		{Name: "Some Other Firefox App", Comment: "Has firefox in comment", Path: "/path/other.desktop", Exec: "other"},
	}

	engine := New(entries)
	results := engine.Search("firefox", entry.AppTypeAll, entries)

	if len(results) < 2 {
		t.Fatal("Expected at least 2 results")
	}

	// Exact match should be first
	if results[0].Name != "Firefox" {
		t.Errorf("First result = %s, want Firefox (exact match)", results[0].Name)
	}

	// Prefix match should be second
	if results[1].Name != "Firefox Developer" {
		t.Errorf("Second result = %s, want Firefox Developer (prefix match)", results[1].Name)
	}
}
