package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/antoniosarro/gofi/internal/domain/entry"
	"github.com/antoniosarro/gofi/internal/favorites"
	"github.com/antoniosarro/gofi/internal/scanner/launchers"
	"github.com/antoniosarro/gofi/internal/search"

	// Import game launchers to trigger registration
	_ "github.com/antoniosarro/gofi/internal/scanner/launchers/heroic"
)

// Scanner finds and manages application entries
type Scanner struct {
	entries           []*entry.Entry
	searchEngine      *search.Engine
	favoritesManager  *favorites.Manager
	scanGameLaunchers bool
}

// NewScanner creates a new scanner
func NewScanner(enableFavorites bool, scanGameLaunchers bool) (*Scanner, error) {
	fm, err := favorites.NewManager(enableFavorites)
	if err != nil {
		return nil, err
	}

	return &Scanner{
		entries:           make([]*entry.Entry, 0),
		favoritesManager:  fm,
		scanGameLaunchers: scanGameLaunchers,
	}, nil
}

// Scan searches for .desktop files and game launcher entries
func (s *Scanner) Scan() error {
	// Track seen entries to avoid duplicates
	seen := make(map[string]bool)

	// Scan .desktop files
	if err := s.scanDesktopFiles(seen); err != nil {
		return fmt.Errorf("scanning desktop files: %w", err)
	}

	// Scan game launchers if enabled
	if s.scanGameLaunchers {
		s.scanGameLauncherEntries(seen)
	}

	// Sort entries: favorites first (by score), then alphabetically.
	// We call this regardless of config, as SortByFavorites (which we just modified)
	// will handle the "disabled" case by sorting alphabetically.
	s.favoritesManager.SortByFavorites(s.entries)

	// Initialize search engine with entries
	s.searchEngine = search.New(s.entries)

	// Cleanup old favorites in background
	if s.favoritesManager != nil {
		go s.favoritesManager.CleanupOldEvents()
	}

	return nil
}

// scanDesktopFiles scans all .desktop file locations
func (s *Scanner) scanDesktopFiles(seen map[string]bool) error {
	searchPaths := SearchPaths()
	existingPaths := FilterExistingPaths(searchPaths)

	for _, searchPath := range existingPaths {
		if err := s.scanDirectory(searchPath, seen); err != nil {
			// Log error but continue with other paths
			if os.Getenv("DEBUG") == "1" {
				fmt.Printf("Error scanning %s: %v\n", searchPath, err)
			}
		}
	}

	return nil
}

// scanDirectory walks a directory and parses .desktop files
func (s *Scanner) scanDirectory(dir string, seen map[string]bool) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip paths we can't access
		}

		// Only process .desktop files
		if info.IsDir() || !strings.HasSuffix(path, ".desktop") {
			return nil
		}

		// Use basename to detect duplicates
		basename := filepath.Base(path)
		if seen[basename] {
			return nil // Skip duplicate
		}

		// Parse the desktop file
		e, err := ParseDesktopFile(path)
		if err != nil || e == nil {
			return nil // Skip invalid or non-application entries
		}

		// Mark as seen and add to entries
		seen[basename] = true
		s.entries = append(s.entries, e)

		return nil
	})
}

// scanGameLauncherEntries scans all registered game launchers
func (s *Scanner) scanGameLauncherEntries(seen map[string]bool) {
	for _, launcher := range launchers.GetAll() {
		entries, err := launcher.Scan()
		if err != nil {
			if os.Getenv("DEBUG") == "1" {
				fmt.Printf("Error scanning %s: %v\n", launcher.Name(), err)
			}
			continue
		}

		// Add entries that haven't been seen
		for _, e := range entries {
			if !seen[e.Path] {
				seen[e.Path] = true
				s.entries = append(s.entries, e)
			}
		}
	}
}

// GetEntries returns all scanned entries
func (s *Scanner) GetEntries() []*entry.Entry {
	return s.entries
}

// GetEntriesByType returns entries filtered by app type
func (s *Scanner) GetEntriesByType(appType entry.AppType) []*entry.Entry {
	if appType == entry.AppTypeAll {
		return s.entries
	}

	filtered := make([]*entry.Entry, 0)
	for _, e := range s.entries {
		if e.GetAppType() == appType {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// Filter searches and filters entries by query and app type
func (s *Scanner) Filter(query string, appType entry.AppType) []*entry.Entry {
	// Safety check: ensure search engine is initialized
	if s.searchEngine == nil {
		// Return all entries if search engine not ready
		if appType == entry.AppTypeAll {
			return s.entries
		}
		// Filter by type manually
		return s.GetEntriesByType(appType)
	}

	// Use search engine for fuzzy matching
	results := s.searchEngine.Search(query, appType, s.entries)

	// Apply favorites sorting if enabled
	if s.favoritesManager != nil {
		s.favoritesManager.SortByFavorites(results)
	}

	// Record search event for non-empty queries (for first result)
	if query != "" && len(results) > 0 && s.favoritesManager != nil {
		s.favoritesManager.RecordSearch(results[0])
	}

	return results
}

// GetAppTypeCounts returns the count of apps for each type
func (s *Scanner) GetAppTypeCounts() map[entry.AppType]int {
	counts := make(map[entry.AppType]int)

	for _, e := range s.entries {
		appType := e.GetAppType()
		counts[appType]++
	}

	counts[entry.AppTypeAll] = len(s.entries)

	return counts
}

// Count returns the total number of entries
func (s *Scanner) Count() int {
	return len(s.entries)
}

// GetFavoritesManager returns the favorites manager
func (s *Scanner) GetFavoritesManager() *favorites.Manager {
	return s.favoritesManager
}
