package favorites

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/antoniosarro/gofi/internal/domain/entry"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	m, err := NewManager(true)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if m == nil {
		t.Fatal("NewManager() returned nil")
	}

	if !m.enabled {
		t.Error("Manager should be enabled")
	}
}

func TestNewManagerDisabled(t *testing.T) {
	m, err := NewManager(false)
	if err != nil {
		t.Fatalf("NewManager(false) error = %v", err)
	}

	if m.enabled {
		t.Error("Manager should be disabled")
	}
}

func TestRecordLaunch(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	m, _ := NewManager(true)
	e := &entry.Entry{
		Name: "Firefox",
		Path: "/test/firefox.desktop",
	}

	m.RecordLaunch(e)

	score := m.GetScore(e)
	if score <= 0 {
		t.Errorf("Score after launch = %v, want > 0", score)
	}
}

func TestIsFavorite(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	m, _ := NewManager(true)
	e := &entry.Entry{
		Name: "Firefox",
		Path: "/test/firefox.desktop",
	}

	// Not a favorite initially
	if m.IsFavorite(e) {
		t.Error("Entry should not be favorite initially")
	}

	// Launch multiple times to become favorite
	for i := 0; i < 10; i++ {
		m.RecordLaunch(e)
	}

	if !m.IsFavorite(e) {
		t.Error("Entry should be favorite after multiple launches")
	}
}

func TestSortByFavorites(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	m, _ := NewManager(true)

	entries := []*entry.Entry{
		{Name: "Zebra", Path: "/test/zebra.desktop"},
		{Name: "Firefox", Path: "/test/firefox.desktop"},
		{Name: "Alacritty", Path: "/test/alacritty.desktop"},
	}

	// Make Firefox a favorite
	for i := 0; i < 10; i++ {
		m.RecordLaunch(entries[1])
	}

	m.SortByFavorites(entries)

	// Firefox should be first (favorite)
	if entries[0].Name != "Firefox" {
		t.Errorf("First entry = %s, want Firefox", entries[0].Name)
	}

	// Others should be alphabetically sorted
	if entries[1].Name != "Alacritty" {
		t.Errorf("Second entry = %s, want Alacritty", entries[1].Name)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "favorites.json")

	// Create manager and record some events
	store1, _ := NewStore(cachePath)
	store1.RecordEvent("/test/firefox.desktop", EventTypeLaunch)
	store1.Save()

	// Create new manager and load
	store2, _ := NewStore(cachePath)
	stats, exists := store2.GetStats("/test/firefox.desktop")

	if !exists {
		t.Fatal("Stats not loaded")
	}

	if len(stats.Events) == 0 {
		t.Error("Events not loaded")
	}
}
