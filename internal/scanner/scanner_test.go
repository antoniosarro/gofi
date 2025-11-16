package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/antoniosarro/gofi/internal/domain/entry"
)

func TestNewScanner(t *testing.T) {
	tests := []struct {
		name              string
		enableFavorites   bool
		scanGameLaunchers bool
		expectFavorites   bool
		expectGameScan    bool
	}{
		{
			name:              "Default options",
			enableFavorites:   false,
			scanGameLaunchers: true,
			expectFavorites:   false,
			expectGameScan:    true,
		},
		{
			name:              "Favorites enabled",
			enableFavorites:   true,
			scanGameLaunchers: false,
			expectFavorites:   true,
			expectGameScan:    false,
		},
		{
			name:              "All enabled",
			enableFavorites:   true,
			scanGameLaunchers: true,
			expectFavorites:   true,
			expectGameScan:    true,
		},
		{
			name:              "All disabled",
			enableFavorites:   false,
			scanGameLaunchers: false,
			expectFavorites:   false,
			expectGameScan:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set temp cache dir for favorites
			tmpDir := t.TempDir()
			os.Setenv("XDG_CACHE_HOME", tmpDir)
			defer os.Unsetenv("XDG_CACHE_HOME")

			s, err := NewScanner(tt.enableFavorites, tt.scanGameLaunchers)
			if err != nil {
				t.Fatalf("NewScanner() error = %v", err)
			}

			if s.scanGameLaunchers != tt.expectGameScan {
				t.Errorf("scanGameLaunchers = %v, want %v", s.scanGameLaunchers, tt.expectGameScan)
			}

			if tt.expectFavorites {
				if s.favoritesManager == nil {
					t.Error("favoritesManager is nil, expected non-nil")
				}
			}
		})
	}
}

func TestParseDesktopFile(t *testing.T) {
	// Create a temporary .desktop file
	tmpDir := t.TempDir()
	desktopFile := filepath.Join(tmpDir, "test.desktop")

	content := `[Desktop Entry]
Type=Application
Name=Test Application
GenericName=Test App
Comment=A test application
Exec=test-app --flag
Icon=test-icon
Terminal=false
Categories=Utility;Development;
`

	if err := os.WriteFile(desktopFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	e, err := ParseDesktopFile(desktopFile)
	if err != nil {
		t.Fatalf("ParseDesktopFile() error = %v", err)
	}

	if e == nil {
		t.Fatal("ParseDesktopFile() returned nil entry")
	}

	if e.Name != "Test Application" {
		t.Errorf("Name = %v, want %v", e.Name, "Test Application")
	}

	if e.GenericName != "Test App" {
		t.Errorf("GenericName = %v, want %v", e.GenericName, "Test App")
	}

	if e.Comment != "A test application" {
		t.Errorf("Comment = %v, want %v", e.Comment, "A test application")
	}

	if e.Exec != "test-app --flag" {
		t.Errorf("Exec = %v, want %v", e.Exec, "test-app --flag")
	}

	if e.Icon != "test-icon" {
		t.Errorf("Icon = %v, want %v", e.Icon, "test-icon")
	}

	if e.Terminal {
		t.Error("Terminal should be false")
	}

	expectedCategories := []string{"Utility", "Development"}
	if len(e.Categories) != len(expectedCategories) {
		t.Errorf("Categories length = %v, want %v", len(e.Categories), len(expectedCategories))
	}
}

func TestParseDesktopFileHidden(t *testing.T) {
	tmpDir := t.TempDir()
	desktopFile := filepath.Join(tmpDir, "hidden.desktop")

	content := `[Desktop Entry]
Type=Application
Name=Hidden App
Exec=hidden-app
NoDisplay=true
`

	if err := os.WriteFile(desktopFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	e, err := ParseDesktopFile(desktopFile)
	if err != nil {
		t.Fatalf("ParseDesktopFile() error = %v", err)
	}

	if e != nil {
		t.Error("ParseDesktopFile() should return nil for hidden entries")
	}
}

func TestParseDesktopFileInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	desktopFile := filepath.Join(tmpDir, "invalid.desktop")

	content := `[Desktop Entry]
Type=Application
Name=Invalid App
# Missing Exec field
`

	if err := os.WriteFile(desktopFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	e, err := ParseDesktopFile(desktopFile)
	if err != nil {
		t.Fatalf("ParseDesktopFile() error = %v", err)
	}

	if e != nil {
		t.Error("ParseDesktopFile() should return nil for invalid entries")
	}
}

func TestGetEntriesByType(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	s, _ := NewScanner(false, false)
	s.entries = []*entry.Entry{
		{Name: "Firefox", Path: "/var/lib/flatpak/exports/share/applications/firefox.desktop"},
		{Name: "Alacritty", Path: "/run/current-system/sw/share/applications/alacritty.desktop"},
		{Name: "Game", Path: "/usr/share/applications/game.desktop", Categories: []string{"Game"}},
	}

	tests := []struct {
		name     string
		appType  entry.AppType
		expected int
	}{
		{"All entries", entry.AppTypeAll, 3},
		{"Flatpak only", entry.AppTypeFlatpak, 1},
		{"NixSystem only", entry.AppTypeNixSystem, 1},
		{"Games only", entry.AppTypeGame, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.GetEntriesByType(tt.appType)
			if len(result) != tt.expected {
				t.Errorf("GetEntriesByType(%v) length = %v, want %v", tt.appType, len(result), tt.expected)
			}
		})
	}
}

func TestGetAppTypeCounts(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	s, _ := NewScanner(false, false)
	s.entries = []*entry.Entry{
		{Name: "Firefox", Path: "/var/lib/flatpak/exports/share/applications/firefox.desktop"},
		{Name: "Chrome", Path: "/var/lib/flatpak/exports/share/applications/chrome.desktop"},
		{Name: "Alacritty", Path: "/run/current-system/sw/share/applications/alacritty.desktop"},
		{Name: "Game", Path: "/usr/share/applications/game.desktop", Categories: []string{"Game"}},
	}

	counts := s.GetAppTypeCounts()

	if counts[entry.AppTypeAll] != 4 {
		t.Errorf("All count = %v, want %v", counts[entry.AppTypeAll], 4)
	}

	if counts[entry.AppTypeFlatpak] != 2 {
		t.Errorf("Flatpak count = %v, want %v", counts[entry.AppTypeFlatpak], 2)
	}

	if counts[entry.AppTypeNixSystem] != 1 {
		t.Errorf("NixSystem count = %v, want %v", counts[entry.AppTypeNixSystem], 1)
	}

	if counts[entry.AppTypeGame] != 1 {
		t.Errorf("Game count = %v, want %v", counts[entry.AppTypeGame], 1)
	}
}

func TestSearchPaths(t *testing.T) {
	paths := SearchPaths()

	if len(paths) == 0 {
		t.Error("SearchPaths() returned no paths")
	}

	// Check that standard paths are included
	expectedPaths := []string{
		"/usr/share/applications",
		"/usr/local/share/applications",
	}

	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if path == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SearchPaths() missing expected path: %s", expected)
		}
	}
}

func TestFilterExistingPaths(t *testing.T) {
	tmpDir := t.TempDir()
	existingPath := filepath.Join(tmpDir, "existing")
	os.MkdirAll(existingPath, 0755)

	nonExistingPath := filepath.Join(tmpDir, "non-existing")

	paths := []string{existingPath, nonExistingPath}
	filtered := FilterExistingPaths(paths)

	if len(filtered) != 1 {
		t.Errorf("FilterExistingPaths() length = %v, want %v", len(filtered), 1)
	}

	if filtered[0] != existingPath {
		t.Errorf("FilterExistingPaths()[0] = %v, want %v", filtered[0], existingPath)
	}
}

func TestScan(t *testing.T) {
	// Create a temporary test environment
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer func() {
		os.Unsetenv("HOME")
		os.Unsetenv("XDG_CACHE_HOME")
	}()

	// Create desktop file directory
	appsDir := filepath.Join(tmpDir, ".local/share/applications")
	os.MkdirAll(appsDir, 0755)

	// Create a test desktop file
	desktopContent := `[Desktop Entry]
Type=Application
Name=Test App
Exec=test-app
Icon=test-icon
Categories=Utility;
`
	desktopPath := filepath.Join(appsDir, "test.desktop")
	os.WriteFile(desktopPath, []byte(desktopContent), 0644)

	// Create scanner and scan
	s, err := NewScanner(false, false)
	if err != nil {
		t.Fatalf("NewScanner() error = %v", err)
	}

	err = s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	entries := s.GetEntries()

	// Should find at least the test desktop file
	if len(entries) == 0 {
		t.Error("Scan() found no entries")
	}

	// Verify test app was found
	foundTest := false
	for _, e := range entries {
		if e.Name == "Test App" {
			foundTest = true
			break
		}
	}

	if !foundTest {
		t.Error("Test App not found in scanned entries")
	}
}

func TestCount(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	s, _ := NewScanner(false, false)
	s.entries = []*entry.Entry{
		{Name: "App1", Path: "/path/app1.desktop"},
		{Name: "App2", Path: "/path/app2.desktop"},
		{Name: "App3", Path: "/path/app3.desktop"},
	}

	count := s.Count()
	if count != 3 {
		t.Errorf("Count() = %d, want 3", count)
	}
}

func TestFilter(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer func() {
		os.Unsetenv("HOME")
		os.Unsetenv("XDG_CACHE_HOME")
	}()

	// Create test entries
	appsDir := filepath.Join(tmpDir, ".local/share/applications")
	os.MkdirAll(appsDir, 0755)

	// Create test desktop files
	apps := []struct {
		name string
		exec string
	}{
		{"Firefox", "firefox"},
		{"Chrome", "chrome"},
		{"Thunderbird", "thunderbird"},
	}

	for _, app := range apps {
		content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Exec=%s
`, app.name, app.exec)
		path := filepath.Join(appsDir, strings.ToLower(app.name)+".desktop")
		os.WriteFile(path, []byte(content), 0644)
	}

	s, err := NewScanner(false, false)
	if err != nil {
		t.Fatalf("NewScanner() error = %v", err)
	}

	err = s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	tests := []struct {
		name      string
		query     string
		appType   entry.AppType
		minResult int
	}{
		{
			name:      "Empty query returns all",
			query:     "",
			appType:   entry.AppTypeAll,
			minResult: 3,
		},
		{
			name:      "Exact match",
			query:     "Firefox",
			appType:   entry.AppTypeAll,
			minResult: 1,
		},
		{
			name:      "Prefix match",
			query:     "fire",
			appType:   entry.AppTypeAll,
			minResult: 1,
		},
		{
			name:      "Fuzzy match",
			query:     "ffx",
			appType:   entry.AppTypeAll,
			minResult: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := s.Filter(tt.query, tt.appType)
			if len(results) < tt.minResult {
				t.Errorf("Filter(%q) returned %d results, want at least %d",
					tt.query, len(results), tt.minResult)
			}
		})
	}
}

func TestFilterWithFavorites(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer func() {
		os.Unsetenv("HOME")
		os.Unsetenv("XDG_CACHE_HOME")
	}()

	// Create test entries
	appsDir := filepath.Join(tmpDir, ".local/share/applications")
	os.MkdirAll(appsDir, 0755)

	content := `[Desktop Entry]
Type=Application
Name=Firefox
Exec=firefox
`
	os.WriteFile(filepath.Join(appsDir, "firefox.desktop"), []byte(content), 0644)

	// Create scanner with favorites enabled
	s, err := NewScanner(true, false)
	if err != nil {
		t.Fatalf("NewScanner() error = %v", err)
	}

	err = s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Make Firefox a favorite
	entries := s.GetEntries()
	if len(entries) > 0 {
		for i := 0; i < 10; i++ {
			s.favoritesManager.RecordLaunch(entries[0])
		}
	}

	// Filter should still work with favorites
	results := s.Filter("", entry.AppTypeAll)
	if len(results) == 0 {
		t.Error("Filter() with favorites returned no results")
	}
}
