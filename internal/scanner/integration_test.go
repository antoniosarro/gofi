package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/antoniosarro/gofi/internal/scanner/launchers"
	_ "github.com/antoniosarro/gofi/internal/scanner/launchers/heroic"
)

func TestScannerIntegration(t *testing.T) {
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

	// Create Heroic game library
	heroicDir := filepath.Join(tmpDir, ".config/heroic/sideload_apps")
	os.MkdirAll(heroicDir, 0755)

	library := map[string]interface{}{
		"games": []map[string]interface{}{
			{
				"runner":       "sideload",
				"app_name":     "test-game",
				"title":        "Test Game",
				"folder_name":  "test-folder",
				"is_installed": true,
			},
		},
	}

	libraryData, _ := json.MarshalIndent(library, "", "  ")
	libraryPath := filepath.Join(heroicDir, "library.json")
	os.WriteFile(libraryPath, libraryData, 0644)

	// Create scanner with game launchers enabled
	s, err := NewScanner(false, true)
	if err != nil {
		t.Fatalf("NewScanner() error = %v", err)
	}

	err = s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	entries := s.GetEntries()

	// Should find both the desktop file and the game
	if len(entries) < 2 {
		t.Errorf("Scan() found %v entries, want at least 2", len(entries))
	}

	// Verify desktop file entry
	foundDesktop := false
	foundGame := false

	for _, e := range entries {
		if e.Name == "Test App" {
			foundDesktop = true
		}
		if e.Name == "Test Game" {
			foundGame = true
		}
	}

	if !foundDesktop {
		t.Error("Desktop file entry not found")
	}

	if !foundGame {
		t.Error("Heroic game entry not found")
	}

	// Test type filtering
	counts := s.GetAppTypeCounts()
	if counts == nil {
		t.Error("GetAppTypeCounts() returned nil")
	}
}

func TestLauncherRegistry(t *testing.T) {
	// Check that Heroic launcher is registered
	launcher, ok := launchers.Get("heroic")
	if !ok {
		t.Error("Heroic launcher not registered")
	}

	if launcher.Name() != "heroic" {
		t.Errorf("Launcher name = %v, want heroic", launcher.Name())
	}

	// Check registry functions
	all := launchers.GetAll()
	if len(all) == 0 {
		t.Error("No launchers registered")
	}

	names := launchers.List()
	if len(names) == 0 {
		t.Error("No launcher names returned")
	}

	count := launchers.Count()
	if count == 0 {
		t.Error("Launcher count is 0")
	}
}

func TestScanWithoutGameLaunchers(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer func() {
		os.Unsetenv("HOME")
		os.Unsetenv("XDG_CACHE_HOME")
	}()

	// Create a desktop file
	appsDir := filepath.Join(tmpDir, ".local/share/applications")
	os.MkdirAll(appsDir, 0755)

	desktopContent := `[Desktop Entry]
Type=Application
Name=Test App
Exec=test-app
`
	desktopPath := filepath.Join(appsDir, "test.desktop")
	os.WriteFile(desktopPath, []byte(desktopContent), 0644)

	// Create Heroic library (should be ignored)
	heroicDir := filepath.Join(tmpDir, ".config/heroic/sideload_apps")
	os.MkdirAll(heroicDir, 0755)

	library := map[string]interface{}{
		"games": []map[string]interface{}{
			{
				"runner":       "sideload",
				"app_name":     "test-game",
				"title":        "Test Game",
				"is_installed": true,
			},
		},
	}

	libraryData, _ := json.MarshalIndent(library, "", "  ")
	os.WriteFile(filepath.Join(heroicDir, "library.json"), libraryData, 0644)

	// Scan with game launchers disabled
	s, err := NewScanner(false, false)
	if err != nil {
		t.Fatalf("NewScanner() error = %v", err)
	}

	err = s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	entries := s.GetEntries()

	// Should only find the desktop file, not the game
	if len(entries) != 1 {
		t.Errorf("Scan() found %v entries, want 1", len(entries))
	}

	if len(entries) > 0 && entries[0].Name != "Test App" {
		t.Errorf("Entry name = %v, want Test App", entries[0].Name)
	}
}

func TestScanWithFavorites(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer func() {
		os.Unsetenv("HOME")
		os.Unsetenv("XDG_CACHE_HOME")
	}()

	// Create a desktop file
	appsDir := filepath.Join(tmpDir, ".local/share/applications")
	os.MkdirAll(appsDir, 0755)

	desktopContent := `[Desktop Entry]
Type=Application
Name=Test App
Exec=test-app
`
	desktopPath := filepath.Join(appsDir, "test.desktop")
	os.WriteFile(desktopPath, []byte(desktopContent), 0644)

	// Create scanner with favorites enabled
	s, err := NewScanner(true, false)
	if err != nil {
		t.Fatalf("NewScanner() error = %v", err)
	}

	if s.favoritesManager == nil {
		t.Fatal("Favorites manager should not be nil when enabled")
	}

	err = s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Test that favorites manager is accessible
	fm := s.GetFavoritesManager()
	if fm == nil {
		t.Error("GetFavoritesManager() returned nil")
	}
}

func TestScanMultiplePaths(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer func() {
		os.Unsetenv("HOME")
		os.Unsetenv("XDG_CACHE_HOME")
	}()

	// Create multiple application directories
	paths := []string{
		filepath.Join(tmpDir, ".local/share/applications"),
		filepath.Join(tmpDir, ".local/share/flatpak/exports/share/applications"),
	}

	for i, path := range paths {
		os.MkdirAll(path, 0755)

		desktopContent := `[Desktop Entry]
Type=Application
Name=Test App ` + string(rune('A'+i)) + `
Exec=test-app-` + string(rune('a'+i)) + `
`
		desktopPath := filepath.Join(path, "test"+string(rune('a'+i))+".desktop")
		os.WriteFile(desktopPath, []byte(desktopContent), 0644)
	}

	s, err := NewScanner(false, false)
	if err != nil {
		t.Fatalf("NewScanner() error = %v", err)
	}

	err = s.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	entries := s.GetEntries()

	// Should find both desktop files
	if len(entries) < 2 {
		t.Errorf("Scan() found %v entries, want at least 2", len(entries))
	}
}
