package fuzzy

import (
	"strings"
	"unicode"
)

// Matcher implements a fuzzy string matching algorithm
type Matcher struct {
	caseSensitive bool
}

// Option is a functional option for Matcher
type Option func(*Matcher)

// WithCaseSensitive sets whether matching is case-sensitive
func WithCaseSensitive(sensitive bool) Option {
	return func(m *Matcher) {
		m.caseSensitive = sensitive
	}
}

// New creates a new fuzzy matcher
func New(opts ...Option) *Matcher {
	m := &Matcher{
		caseSensitive: false, // Default: case-insensitive
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Result represents a fuzzy match result
type Result struct {
	Text           string
	Score          int
	MatchedIndices []int
}

// Match performs fuzzy matching and returns a result with score
// Algorithm inspired by Sublime Text's fuzzy matching:
// 1. Sequential character matching (all pattern chars must appear in order)
// 2. Bonus for consecutive matches
// 3. Bonus for camelCase/snake_case boundaries
// 4. Bonus for start of word matches
// 5. Penalty for gaps between matches
func (m *Matcher) Match(pattern, text string) *Result {
	if pattern == "" {
		return &Result{Text: text, Score: 0, MatchedIndices: []int{}}
	}

	// Normalize for case-insensitive matching
	searchPattern := pattern
	searchText := text
	if !m.caseSensitive {
		searchPattern = strings.ToLower(pattern)
		searchText = strings.ToLower(text)
	}

	// Check if all characters in pattern exist in text (in order)
	matchedIndices := m.findMatches(searchPattern, searchText)
	if matchedIndices == nil {
		return nil // No match
	}

	// Calculate score based on match quality
	score := m.calculateScore(searchPattern, searchText, matchedIndices)

	return &Result{
		Text:           text,
		Score:          score,
		MatchedIndices: matchedIndices,
	}
}

// findMatches finds the indices of matched characters
func (m *Matcher) findMatches(pattern, text string) []int {
	patternIdx := 0
	textIdx := 0
	matchedIndices := make([]int, 0, len(pattern))

	for textIdx < len(text) && patternIdx < len(pattern) {
		if text[textIdx] == pattern[patternIdx] {
			matchedIndices = append(matchedIndices, textIdx)
			patternIdx++
		}
		textIdx++
	}

	// No match if we didn't find all pattern characters
	if patternIdx < len(pattern) {
		return nil
	}

	return matchedIndices
}

// calculateScore computes match quality score
func (m *Matcher) calculateScore(pattern, text string, matchedIndices []int) int {
	if len(matchedIndices) == 0 {
		return 0
	}

	score := 0
	consecutive := 0

	// Base score for matching
	score += len(pattern) * 10

	for i, idx := range matchedIndices {
		// Bonus for match at start
		if idx == 0 {
			score += 15
		}

		// Bonus for consecutive characters
		if i > 0 && matchedIndices[i-1] == idx-1 {
			consecutive++
			score += 15 + (consecutive * 5) // Increasing bonus for longer sequences
		} else {
			consecutive = 0
		}

		// Bonus for matching after separator or boundary
		if idx > 0 {
			prevChar := rune(text[idx-1])
			currChar := rune(text[idx])

			// Word boundary bonuses
			if isSeparator(prevChar) {
				score += 20 // After space, dash, underscore, etc.
			} else if isLower(prevChar) && isUpper(currChar) {
				score += 15 // CamelCase boundary
			} else if unicode.IsDigit(prevChar) && !unicode.IsDigit(currChar) {
				score += 10 // Number to letter transition
			}
		}

		// Penalty for gaps
		if i > 0 {
			gap := idx - matchedIndices[i-1] - 1
			if gap > 0 {
				score -= gap * 2 // Penalize gaps between matches
			}
		}
	}

	// Bonus for matching a higher percentage of the text
	matchPercentage := float64(len(pattern)) / float64(len(text))
	score += int(matchPercentage * 50)

	// Penalty for very long text compared to pattern
	lengthRatio := float64(len(text)) / float64(len(pattern))
	if lengthRatio > 3 {
		score -= int((lengthRatio - 3) * 5)
	}

	return score
}

// isSeparator checks if a character is a word separator
func isSeparator(r rune) bool {
	return unicode.IsSpace(r) || r == '-' || r == '_' || r == '/' || r == '.' || r == ':'
}

// isLower checks if a rune is lowercase
func isLower(r rune) bool {
	return unicode.IsLower(r)
}

// isUpper checks if a rune is uppercase
func isUpper(r rune) bool {
	return unicode.IsUpper(r)
}
