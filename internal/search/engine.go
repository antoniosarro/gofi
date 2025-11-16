package search

import (
	"sort"
	"strings"

	"github.com/antoniosarro/gofi/internal/domain/entry"
	"github.com/antoniosarro/gofi/internal/search/fuzzy"
)

const (
	// MinimumScore is the threshold below which results are filtered out
	MinimumScore = 50
)

// MatchType represents the type of match
type MatchType int

const (
	ExactMatch MatchType = iota
	PrefixMatch
	FuzzyMatch
	ContainsMatch
	TokenMatch
)

// ScoredEntry represents an entry with its match score
type ScoredEntry struct {
	Entry     *entry.Entry
	Score     int
	MatchType MatchType
}

// Engine provides optimized search with caching and fuzzy matching
type Engine struct {
	indexer      *Indexer
	fuzzyMatcher *fuzzy.Matcher
}

// Option is a functional option for Engine
type Option func(*Engine)

// WithFuzzyMatcher sets a custom fuzzy matcher
func WithFuzzyMatcher(matcher *fuzzy.Matcher) Option {
	return func(e *Engine) {
		e.fuzzyMatcher = matcher
	}
}

// New creates a new search engine
func New(entries []*entry.Entry, opts ...Option) *Engine {
	e := &Engine{
		indexer:      NewIndexer(),
		fuzzyMatcher: fuzzy.New(),
	}

	// Apply options
	for _, opt := range opts {
		opt(e)
	}

	// Build indices
	e.indexer.Build(entries)

	return e
}

// Search performs optimized search with ranking and score filtering
func (e *Engine) Search(query string, appType entry.AppType, entries []*entry.Entry) []*entry.Entry {
	if query == "" {
		// Return all entries of the specified type
		return filterByType(entries, appType)
	}

	queryLower := strings.ToLower(query)
	queryTokens := tokenize(query)

	scored := make([]ScoredEntry, 0)

	for _, ent := range entries {
		// Filter by app type first
		if appType != entry.AppTypeAll && ent.GetAppType() != appType {
			continue
		}

		index := e.indexer.Get(ent.Path)
		if index == nil {
			// Entry not indexed, skip
			continue
		}

		score, matchType := e.scoreEntry(query, queryLower, queryTokens, index)

		// Filter out results below minimum score threshold
		if score < MinimumScore {
			continue
		}

		scored = append(scored, ScoredEntry{
			Entry:     ent,
			Score:     score,
			MatchType: matchType,
		})
	}

	// Sort by match type first, then by score
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].MatchType != scored[j].MatchType {
			return scored[i].MatchType < scored[j].MatchType
		}
		return scored[i].Score > scored[j].Score
	})

	// Extract entries
	result := make([]*entry.Entry, len(scored))
	for i := range scored {
		result[i] = scored[i].Entry
	}

	return result
}

// scoreEntry calculates the score and match type for an entry
func (e *Engine) scoreEntry(query, queryLower string, queryTokens []string, index *Index) (int, MatchType) {
	score := 0
	matchType := TokenMatch

	// 1. Exact match on name (highest priority)
	if index.NameNormalized == queryLower {
		score = 1000
		matchType = ExactMatch
		return score, matchType
	}

	// 2. Prefix match on name
	if strings.HasPrefix(index.NameNormalized, queryLower) {
		score = 800
		matchType = PrefixMatch
		return score, matchType
	}

	// 3. Fuzzy match on searchable text
	fuzzyResult := e.fuzzyMatcher.Match(query, index.SearchableText)
	if fuzzyResult != nil {
		score = fuzzyResult.Score
		matchType = FuzzyMatch
	} else {
		// 4. Fallback to contains match
		if strings.Contains(index.NameNormalized, queryLower) {
			score = 400
			matchType = ContainsMatch
		} else if strings.Contains(strings.ToLower(index.Entry.Comment), queryLower) {
			score = 200
			matchType = ContainsMatch
		} else if matchTokens(queryTokens, index.CommentTokens) {
			// 5. Token-based match
			score = 100
			matchType = TokenMatch
		} else {
			// No match
			return 0, TokenMatch
		}
	}

	// Bonus for generic name match
	if index.Entry.GenericName != "" {
		if strings.Contains(strings.ToLower(index.Entry.GenericName), queryLower) {
			score += 100
		}
	}

	// Bonus for category match
	for _, cat := range index.CategoryTokens {
		if strings.Contains(strings.ToLower(cat), queryLower) {
			score += 50
			break
		}
	}

	return score, matchType
}

// UpdateIndex updates the search index for a new/modified entry
func (e *Engine) UpdateIndex(ent *entry.Entry) {
	e.indexer.Add(ent)
}

// RemoveIndex removes an entry from the index
func (e *Engine) RemoveIndex(path string) {
	e.indexer.Remove(path)
}

// matchTokens checks if any query tokens match target tokens
func matchTokens(queryTokens, targetTokens []string) bool {
	for _, qt := range queryTokens {
		for _, tt := range targetTokens {
			if strings.Contains(strings.ToLower(tt), strings.ToLower(qt)) {
				return true
			}
		}
	}
	return false
}

// filterByType filters entries by app type
func filterByType(entries []*entry.Entry, appType entry.AppType) []*entry.Entry {
	if appType == entry.AppTypeAll {
		return entries
	}

	filtered := make([]*entry.Entry, 0)
	for _, e := range entries {
		if e.GetAppType() == appType {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
