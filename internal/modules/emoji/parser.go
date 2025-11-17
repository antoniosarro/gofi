package emoji

import (
	"bufio"
	"os"
	"strings"
)

// ParseEmojiFile reads and parses the all_emojis.txt file
func ParseEmojiFile(path string) ([]Emoji, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var emojis []Emoji
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		emoji := parseEmojiLine(line)
		if emoji != nil {
			emojis = append(emojis, *emoji)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return emojis, nil
}

// parseEmojiLine parses a single line from the emoji file
// Format: 😁    Smileys & People        face-positive   beaming face with smiling eyes	beaming face with smiling eyes | eye | face | grin | smile
func parseEmojiLine(line string) *Emoji {
	// Split by tabs
	parts := strings.Split(line, "\t")
	if len(parts) < 5 {
		return nil
	}

	char := strings.TrimSpace(parts[0])
	category := strings.TrimSpace(parts[1])
	subcategory := strings.TrimSpace(parts[2])
	name := strings.TrimSpace(parts[3])
	keywordsStr := strings.TrimSpace(parts[4])

	// Parse keywords (separated by |)
	var keywords []string
	for _, kw := range strings.Split(keywordsStr, "|") {
		keyword := strings.TrimSpace(kw)
		if keyword != "" {
			keywords = append(keywords, keyword)
		}
	}

	return &Emoji{
		Char:        char,
		Name:        name,
		Category:    category,
		Subcategory: subcategory,
		Keywords:    keywords,
	}
}

// GetCategories returns unique categories from emoji list
func GetCategories(emojis []Emoji) []string {
	categoryMap := make(map[string]bool)
	var categories []string

	for _, emoji := range emojis {
		if !categoryMap[emoji.Category] {
			categoryMap[emoji.Category] = true
			categories = append(categories, emoji.Category)
		}
	}

	return categories
}

// FilterByCategory filters emojis by category
func FilterByCategory(emojis []Emoji, category string) []Emoji {
	if category == "" || category == string(CategoryAll) {
		return emojis
	}

	filtered := make([]Emoji, 0)
	for _, emoji := range emojis {
		if emoji.Category == category {
			filtered = append(filtered, emoji)
		}
	}

	return filtered
}
