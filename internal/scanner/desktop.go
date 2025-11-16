package scanner

import (
	"bufio"
	"os"
	"strings"

	"github.com/antoniosarro/gofi/internal/domain/entry"
)

// ParseDesktopFile parses a .desktop file according to freedesktop.org specification
func ParseDesktopFile(path string) (*entry.Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	e := &entry.Entry{Path: path}
	scanner := bufio.NewScanner(file)
	inDesktopEntry := false
	noDisplay := false
	hidden := false
	isApplication := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Track when we're in the [Desktop Entry] section
		if line == "[Desktop Entry]" {
			inDesktopEntry = true
			continue
		}

		// Exit [Desktop Entry] section when we hit another section
		if strings.HasPrefix(line, "[") {
			inDesktopEntry = false
			continue
		}

		// Skip lines outside [Desktop Entry] section, empty lines, and comments
		if !inDesktopEntry || line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value pairs
		key, value, ok := parseKeyValue(line)
		if !ok {
			continue
		}

		// Extract relevant fields
		switch key {
		case "Type":
			isApplication = (value == "Application")
		case "Name":
			e.Name = value
		case "GenericName":
			e.GenericName = value
		case "Comment":
			e.Comment = value
		case "Exec":
			e.Exec = value
		case "Icon":
			e.Icon = value
		case "Terminal":
			e.Terminal = (value == "true")
		case "Categories":
			e.Categories = parseCategories(value)
		case "NoDisplay":
			noDisplay = (value == "true")
		case "Hidden":
			hidden = (value == "true")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Only process Application type entries
	if !isApplication {
		return nil, nil
	}

	// Filter out hidden or no-display entries
	if noDisplay || hidden {
		return nil, nil
	}

	// Validate the entry
	if err := e.Validate(); err != nil {
		return nil, nil // Skip invalid entries
	}

	return e, nil
}

// parseKeyValue splits a line into key and value
func parseKeyValue(line string) (key, value string, ok bool) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
}

// parseCategories parses the semicolon-separated categories
func parseCategories(value string) []string {
	var categories []string
	for _, cat := range strings.Split(value, ";") {
		cat = strings.TrimSpace(cat)
		if cat != "" {
			categories = append(categories, cat)
		}
	}
	return categories
}
