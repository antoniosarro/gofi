package favorites

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Store handles persistence of favorites data
type Store struct {
	cachePath string
	stats     map[string]*AppStats
	mu        sync.RWMutex
	dirty     bool // Tracks if data needs to be saved
}

// NewStore creates a new favorites store
func NewStore(cachePath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return nil, err
	}

	store := &Store{
		cachePath: cachePath,
		stats:     make(map[string]*AppStats),
		dirty:     false,
	}

	// Load existing data
	if err := store.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}

// Load reads favorites from cache
func (s *Store) Load() error {
	data, err := os.ReadFile(s.cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No cache yet
		}
		return err
	}

	var statsList []*AppStats
	if err := json.Unmarshal(data, &statsList); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Convert to map
	for _, stats := range statsList {
		s.stats[stats.DesktopFile] = stats
	}

	return nil
}

// Save writes favorites to cache
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Only save if data has changed
	if !s.dirty {
		return nil
	}

	// Convert map to slice
	statsList := make([]*AppStats, 0, len(s.stats))
	for _, stats := range s.stats {
		statsList = append(statsList, stats)
	}

	data, err := json.MarshalIndent(statsList, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(s.cachePath, data, 0644)
	if err != nil {
		return err
	}

	// Clear dirty flag after successful save
	s.dirty = false
	return nil
}

// GetStats retrieves stats for a desktop file
func (s *Store) GetStats(desktopFile string) (*AppStats, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats, exists := s.stats[desktopFile]
	return stats, exists
}

// RecordEvent adds an event and marks data as dirty
func (s *Store) RecordEvent(desktopFile string, eventType EventType) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats, exists := s.stats[desktopFile]
	if !exists {
		stats = &AppStats{
			DesktopFile: desktopFile,
			Events:      make([]Event, 0),
		}
		s.stats[desktopFile] = stats
	}

	// Add new event
	stats.Events = append(stats.Events, Event{
		Timestamp: time.Now(),
		Type:      eventType,
	})

	// Prune old events if we exceed max
	if len(stats.Events) > MaxEvents {
		stats.Events = stats.Events[len(stats.Events)-MaxEvents:]
	}

	// Mark as dirty
	s.dirty = true
}

// CleanupOldEvents removes events older than 90 days
func (s *Store) CleanupOldEvents() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -90)
	needsSave := false

	for desktopFile, stats := range s.stats {
		filtered := make([]Event, 0)
		for _, event := range stats.Events {
			if event.Timestamp.After(cutoff) {
				filtered = append(filtered, event)
			}
		}

		if len(filtered) == 0 {
			// Remove entry if no recent events
			delete(s.stats, desktopFile)
			needsSave = true
		} else if len(filtered) != len(stats.Events) {
			stats.Events = filtered
			needsSave = true
		}
	}

	if needsSave {
		s.dirty = true
	}
}

// GetAllStats returns a copy of all stats (for iteration)
func (s *Store) GetAllStats() map[string]*AppStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*AppStats, len(s.stats))
	for k, v := range s.stats {
		result[k] = v
	}
	return result
}

// getCacheDir returns the cache directory path
func getCacheDir() string {
	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		cacheHome = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return filepath.Join(cacheHome, "gofi")
}
