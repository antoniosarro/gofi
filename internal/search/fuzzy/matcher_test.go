package fuzzy

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name              string
		opts              []Option
		expectedSensitive bool
	}{
		{
			name:              "Default case-insensitive",
			opts:              nil,
			expectedSensitive: false,
		},
		{
			name:              "Case-sensitive",
			opts:              []Option{WithCaseSensitive(true)},
			expectedSensitive: true,
		},
		{
			name:              "Case-insensitive explicit",
			opts:              []Option{WithCaseSensitive(false)},
			expectedSensitive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.opts...)
			if m.caseSensitive != tt.expectedSensitive {
				t.Errorf("caseSensitive = %v, want %v", m.caseSensitive, tt.expectedSensitive)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		text      string
		wantMatch bool
		minScore  int // Minimum expected score
	}{
		{
			name:      "Exact match",
			pattern:   "firefox",
			text:      "firefox",
			wantMatch: true,
			minScore:  100,
		},
		{
			name:      "Prefix match",
			pattern:   "fire",
			text:      "firefox",
			wantMatch: true,
			minScore:  80,
		},
		{
			name:      "Scattered match",
			pattern:   "ffx",
			text:      "firefox",
			wantMatch: true,
			minScore:  30,
		},
		{
			name:      "CamelCase match",
			pattern:   "fb",
			text:      "FooBar",
			wantMatch: true,
			minScore:  40,
		},
		{
			name:      "No match",
			pattern:   "xyz",
			text:      "firefox",
			wantMatch: false,
		},
		{
			name:      "Empty pattern",
			pattern:   "",
			text:      "firefox",
			wantMatch: true,
			minScore:  0,
		},
		{
			name:      "Case insensitive match",
			pattern:   "FIRE",
			text:      "firefox",
			wantMatch: true,
			minScore:  80,
		},
	}

	m := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.Match(tt.pattern, tt.text)

			if tt.wantMatch {
				if result == nil {
					t.Error("Match() returned nil, want match")
					return
				}
				if result.Score < tt.minScore {
					t.Errorf("Match() score = %v, want >= %v", result.Score, tt.minScore)
				}
				if result.Text != tt.text {
					t.Errorf("Match() text = %v, want %v", result.Text, tt.text)
				}
			} else {
				if result != nil {
					t.Errorf("Match() = %+v, want nil", result)
				}
			}
		})
	}
}

func TestMatchCaseSensitive(t *testing.T) {
	m := New(WithCaseSensitive(true))

	result := m.Match("Fire", "firefox")
	if result != nil {
		t.Error("Case-sensitive match should fail for different cases")
	}

	result = m.Match("fire", "firefox")
	if result == nil {
		t.Error("Case-sensitive match should succeed for matching cases")
	}
}

func TestFindMatches(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		text     string
		expected []int
	}{
		{
			name:     "Sequential match",
			pattern:  "abc",
			text:     "abc",
			expected: []int{0, 1, 2},
		},
		{
			name:     "Scattered match",
			pattern:  "ac",
			text:     "abc",
			expected: []int{0, 2},
		},
		{
			name:     "No match",
			pattern:  "xyz",
			text:     "abc",
			expected: nil,
		},
		{
			name:     "Partial pattern",
			pattern:  "ab",
			text:     "aabbcc",
			expected: []int{0, 2}, // First occurrence
		},
	}

	m := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.findMatches(tt.pattern, tt.text)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("findMatches() = %v, want nil", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("findMatches() length = %v, want %v", len(result), len(tt.expected))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("findMatches()[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestCalculateScore(t *testing.T) {
	m := New()

	tests := []struct {
		name     string
		pattern  string
		text     string
		indices  []int
		minScore int
	}{
		{
			name:     "Match at start",
			pattern:  "fir",
			text:     "firefox",
			indices:  []int{0, 1, 2},
			minScore: 100,
		},
		{
			name:     "Scattered match",
			pattern:  "ffx",
			text:     "firefox",
			indices:  []int{0, 4, 5},
			minScore: 30,
		},
		{
			name:     "CamelCase boundary",
			pattern:  "fb",
			text:     "FooBar",
			indices:  []int{0, 3},
			minScore: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := m.calculateScore(tt.pattern, tt.text, tt.indices)
			if score < tt.minScore {
				t.Errorf("calculateScore() = %v, want >= %v", score, tt.minScore)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{
			name:     "Identical strings",
			s1:       "kitten",
			s2:       "kitten",
			expected: 0,
		},
		{
			name:     "One substitution",
			s1:       "kitten",
			s2:       "sitten",
			expected: 1,
		},
		{
			name:     "Classic example",
			s1:       "kitten",
			s2:       "sitting",
			expected: 3,
		},
		{
			name:     "Empty strings",
			s1:       "",
			s2:       "",
			expected: 0,
		},
		{
			name:     "One empty",
			s1:       "abc",
			s2:       "",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LevenshteinDistance(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("LevenshteinDistance(%q, %q) = %v, want %v", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}

func BenchmarkMatch(b *testing.B) {
	m := New()
	pattern := "fire"
	text := "firefox web browser"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Match(pattern, text)
	}
}

func BenchmarkLevenshteinDistance(b *testing.B) {
	s1 := "kitten"
	s2 := "sitting"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LevenshteinDistance(s1, s2)
	}
}
