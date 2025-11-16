package favorites

import (
	"path/filepath"
	"sort"

	"github.com/antoniosarro/gofi/internal/domain/entry"
)

// Manager handles favorite app tracking and scoring
type Manager struct {
	store   *Store
	scorer  *Scorer
	enabled bool
}

// NewManager creates a new favorites manager
func NewManager(enabled bool) (*Manager, error) {
	if !enabled {
		return &Manager{
			enabled: false,
		}, nil
	}

	cachePath := filepath.Join(getCacheDir(), "favorites.json")
	store, err := NewStore(cachePath)
	if err != nil {
		return nil, err
	}

	return &Manager{
		store:   store,
		scorer:  NewScorer(),
		enabled: enabled,
	}, nil
}

// RecordLaunch records an app launch event
func (m *Manager) RecordLaunch(e *entry.Entry) {
	if !m.enabled {
		return
	}
	m.store.RecordEvent(e.Path, EventTypeLaunch)
}

// RecordSearch records a search event
func (m *Manager) RecordSearch(e *entry.Entry) {
	if !m.enabled {
		return
	}
	m.store.RecordEvent(e.Path, EventTypeSearch)
}

// GetScore returns the score for an entry
func (m *Manager) GetScore(e *entry.Entry) float64 {
	if !m.enabled {
		return 0
	}

	stats, exists := m.store.GetStats(e.Path)
	if !exists {
		return 0
	}

	return m.scorer.CalculateScore(stats)
}

// IsFavorite checks if an app is a favorite
func (m *Manager) IsFavorite(e *entry.Entry) bool {
	if !m.enabled {
		return false
	}
	score := m.GetScore(e)
	return m.scorer.IsFavorite(score)
}

// SortByFavorites sorts entries with favorites first, then alphabetically
func (m *Manager) SortByFavorites(entries []*entry.Entry) {
	if !m.enabled {
		// If favorites are disabled, just sort alphabetically
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Name < entries[j].Name
		})
		return
	}

	// Calculate scores for all entries
	scores := make(map[string]float64)
	for _, e := range entries {
		scores[e.Path] = m.GetScore(e)
	}

	// Sort: favorites (by score desc) first, then non-favorites (alphabetically)
	sort.Slice(entries, func(i, j int) bool {
		scoreI := scores[entries[i].Path]
		scoreJ := scores[entries[j].Path]

		isFavI := m.scorer.IsFavorite(scoreI)
		isFavJ := m.scorer.IsFavorite(scoreJ)

		// Both are favorites: sort by score (higher first)
		if isFavI && isFavJ {
			return scoreI > scoreJ
		}

		// One is favorite: favorite comes first
		if isFavI != isFavJ {
			return isFavI
		}

		// Neither is favorite: sort alphabetically
		return entries[i].Name < entries[j].Name
	})
}

// Save writes favorites to cache (synchronous - call before app exits)
func (m *Manager) Save() error {
	if !m.enabled {
		return nil
	}
	return m.store.Save()
}

// CleanupOldEvents removes events older than 90 days
func (m *Manager) CleanupOldEvents() {
	if !m.enabled {
		return
	}
	m.store.CleanupOldEvents()
}
