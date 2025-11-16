package entry

import (
	"os"
	"strings"
	"time"
)

// Entry represents a desktop application
type Entry struct {
	Name        string
	GenericName string
	Comment     string
	Exec        string
	Icon        string
	Terminal    bool
	Categories  []string
	Path        string // Path to .desktop file or unique identifier
	LastUsed    time.Time
}

// GetAppType determines the type/source of the application based on its path
func (e *Entry) GetAppType() AppType {
	// Check if it's a Flatpak
	if strings.Contains(e.Path, "flatpak") {
		return AppTypeFlatpak
	}

	// Check if it's from Nix
	if e.isNixApp() {
		return e.getNixAppType()
	}

	// Check if it's a game based on categories
	if e.isGame() {
		return AppTypeGame
	}

	// Check if it's a system app (traditional Linux paths)
	if e.isSystemApp() {
		return AppTypeSystem
	}

	return AppTypeOther
}

// isNixApp checks if the application is from Nix
func (e *Entry) isNixApp() bool {
	return strings.Contains(e.Path, "/nix/store") ||
		strings.Contains(e.Path, ".nix-profile") ||
		strings.Contains(e.Path, "/run/current-system") ||
		strings.Contains(e.Path, "/etc/profiles/per-user") ||
		strings.Contains(e.Exec, "/nix/store")
}

// getNixAppType distinguishes between system and home-manager Nix apps
func (e *Entry) getNixAppType() AppType {
	homeDir := os.Getenv("HOME")
	userName := os.Getenv("USER")

	// Home-manager paths
	if strings.Contains(e.Path, homeDir+"/.nix-profile") ||
		strings.Contains(e.Path, homeDir+"/.local/state/nix/profile") ||
		strings.Contains(e.Path, "/etc/profiles/per-user/"+userName) {
		return AppTypeNixHome
	}

	// System-level Nix paths
	if strings.Contains(e.Path, "/run/current-system") ||
		strings.Contains(e.Path, "/nix/var/nix/profiles/default") {
		return AppTypeNixSystem
	}

	// Fallback: if it's in nix store but we can't determine, assume system
	return AppTypeNixSystem
}

// isGame checks if the application is a game based on categories
func (e *Entry) isGame() bool {
	for _, cat := range e.Categories {
		catLower := strings.ToLower(cat)
		if catLower == "game" || catLower == "games" {
			return true
		}
	}
	return false
}

// isSystemApp checks if it's a traditional system application
func (e *Entry) isSystemApp() bool {
	return strings.HasPrefix(e.Path, "/usr/share/applications") ||
		strings.HasPrefix(e.Path, "/usr/local/share/applications")
}

// Match returns a score for how well this entry matches the query
// This is a placeholder - actual matching is handled by the search engine
func (e *Entry) Match(query string) int {
	if query == "" {
		return 1
	}
	// Basic scoring - real implementation is in search package
	return 0
}

// Clone creates a deep copy of the entry
func (e *Entry) Clone() *Entry {
	categories := make([]string, len(e.Categories))
	copy(categories, e.Categories)

	return &Entry{
		Name:        e.Name,
		GenericName: e.GenericName,
		Comment:     e.Comment,
		Exec:        e.Exec,
		Icon:        e.Icon,
		Terminal:    e.Terminal,
		Categories:  categories,
		Path:        e.Path,
		LastUsed:    e.LastUsed,
	}
}

// Validate checks if the entry has required fields
func (e *Entry) Validate() error {
	if e.Name == "" {
		return ErrMissingName
	}
	if e.Exec == "" {
		return ErrMissingExec
	}
	return nil
}
