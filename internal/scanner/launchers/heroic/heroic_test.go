package heroic

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLauncherName(t *testing.T) {
	l := &Launcher{}
	if l.Name() != "heroic" {
		t.Errorf("Name() = %v, want %v", l.Name(), "heroic")
	}
}

func TestScanNoLibrary(t *testing.T) {
	// Set a temporary home directory
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	l := &Launcher{}
	entries, err := l.Scan()

	if err != nil {
		t.Errorf("Scan() error = %v, want nil", err)
	}

	if len(entries) > 0 {
		t.Errorf("Scan() returned entries when library doesn't exist")
	}
}

func TestScanWithLibrary(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	heroicDir := filepath.Join(tmpDir, ".config", "heroic", "sideload_apps")
	os.MkdirAll(heroicDir, 0755)

	// Create a test library
	library := Library{
		Games: []Game{
			{
				Runner:      "sideload",
				AppName:     "test-game",
				Title:       "Test Game",
				FolderName:  "test-game-folder",
				IsInstalled: true,
			},
			{
				Runner:      "sideload",
				AppName:     "not-installed",
				Title:       "Not Installed Game",
				IsInstalled: false,
			},
		},
	}

	libraryData, _ := json.MarshalIndent(library, "", "  ")
	libraryPath := filepath.Join(heroicDir, "library.json")
	os.WriteFile(libraryPath, libraryData, 0644)

	l := &Launcher{}
	entries, err := l.Scan()

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Scan() returned %v entries, want 1", len(entries))
	}

	if len(entries) > 0 {
		e := entries[0]
		if e.Name != "Test Game" {
			t.Errorf("Entry Name = %v, want %v", e.Name, "Test Game")
		}

		expectedExec := "xdg-open heroic://launch/sideload/test-game"
		if e.Exec != expectedExec {
			t.Errorf("Entry Exec = %v, want %v", e.Exec, expectedExec)
		}

		if len(e.Categories) == 0 || e.Categories[0] != "Game" {
			t.Errorf("Entry Categories = %v, want [Game]", e.Categories)
		}
	}
}

func TestParseLibrary(t *testing.T) {
	tmpDir := t.TempDir()
	libraryPath := filepath.Join(tmpDir, "library.json")

	library := Library{
		Games: []Game{
			{
				Runner:      "sideload",
				AppName:     "test-game",
				Title:       "Test Game",
				FolderName:  "test-folder",
				IsInstalled: true,
			},
		},
	}

	data, _ := json.MarshalIndent(library, "", "  ")
	os.WriteFile(libraryPath, data, 0644)

	l := &Launcher{}
	parsed, err := l.parseLibrary(libraryPath)

	if err != nil {
		t.Fatalf("parseLibrary() error = %v", err)
	}

	if len(parsed.Games) != 1 {
		t.Errorf("parseLibrary() games count = %v, want 1", len(parsed.Games))
	}

	if parsed.Games[0].Title != "Test Game" {
		t.Errorf("parseLibrary() game title = %v, want Test Game", parsed.Games[0].Title)
	}
}

func TestLoadGameCategories(t *testing.T) {
	tmpDir := t.TempDir()
	gamesConfigDir := filepath.Join(tmpDir, "GamesConfig")
	os.MkdirAll(gamesConfigDir, 0755)

	// Create a test game config
	gameConfig := map[string]interface{}{
		"version": "1.0",
		"test-game": map[string]interface{}{
			"categories": []string{"Action", "Adventure"},
		},
	}

	configData, _ := json.MarshalIndent(gameConfig, "", "  ")
	configPath := filepath.Join(gamesConfigDir, "test-game.json")
	os.WriteFile(configPath, configData, 0644)

	l := &Launcher{}
	categories := l.loadGameCategories(tmpDir, "test-game")

	if len(categories) != 2 {
		t.Errorf("loadGameCategories() returned %v categories, want 2", len(categories))
	}

	expected := []string{"Action", "Adventure"}
	for i, cat := range categories {
		if cat != expected[i] {
			t.Errorf("categories[%d] = %v, want %v", i, cat, expected[i])
		}
	}
}

func TestLoadGameCategoriesNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	l := &Launcher{}
	categories := l.loadGameCategories(tmpDir, "non-existent")

	if categories != nil {
		t.Errorf("loadGameCategories() = %v, want nil", categories)
	}
}

func TestGameToEntry(t *testing.T) {
	game := Game{
		Runner:     "sideload",
		AppName:    "test-game",
		Title:      "Test Game",
		FolderName: "test-folder",
	}

	l := &Launcher{}
	entry := l.gameToEntry(game, "/tmp/heroic")

	if entry.Name != "Test Game" {
		t.Errorf("Entry.Name = %v, want Test Game", entry.Name)
	}

	if entry.Exec != "xdg-open heroic://launch/sideload/test-game" {
		t.Errorf("Entry.Exec = %v, want xdg-open heroic://launch/sideload/test-game", entry.Exec)
	}

	if entry.Icon != "applications-games" {
		t.Errorf("Entry.Icon = %v, want applications-games", entry.Icon)
	}

	if entry.Path != "heroic-sideload-test-game" {
		t.Errorf("Entry.Path = %v, want heroic-sideload-test-game", entry.Path)
	}

	if len(entry.Categories) == 0 || entry.Categories[0] != "Game" {
		t.Errorf("Entry.Categories = %v, want [Game]", entry.Categories)
	}
}
