package emoji

import (
	"strings"
)

// Search searches emojis by query matching name and keywords
func Search(emojis []Emoji, query string) []Emoji {
	if query == "" {
		return emojis
	}

	query = strings.ToLower(strings.TrimSpace(query))
	results := make([]Emoji, 0)

	for _, emoji := range emojis {
		if matchesQuery(emoji, query) {
			results = append(results, emoji)
		}
	}

	return results
}

func matchesQuery(emoji Emoji, query string) bool {
	// Check name
	if strings.Contains(strings.ToLower(emoji.Name), query) {
		return true
	}

	// Check keywords
	for _, keyword := range emoji.Keywords {
		if strings.Contains(strings.ToLower(keyword), query) {
			return true
		}
	}

	// Check category
	if strings.Contains(strings.ToLower(emoji.Category), query) {
		return true
	}

	return false
}
