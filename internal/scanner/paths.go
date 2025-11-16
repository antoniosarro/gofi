package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// SearchPaths returns all directories to search for .desktop files
func SearchPaths() []string {
	homeDir := os.Getenv("HOME")
	userName := os.Getenv("USER")

	paths := []string{
		// Standard Linux paths
		"/usr/share/applications",
		"/usr/local/share/applications",
		filepath.Join(homeDir, ".local/share/applications"),

		// Flatpak paths
		"/var/lib/flatpak/exports/share/applications",
		filepath.Join(homeDir, ".local/share/flatpak/exports/share/applications"),

		// NixOS system profile
		"/run/current-system/sw/share/applications",
		"/nix/var/nix/profiles/default/share/applications",

		// NixOS user profiles (home-manager)
		filepath.Join(homeDir, ".nix-profile/share/applications"),
		filepath.Join(homeDir, ".local/state/nix/profiles/profile/share/applications"),
		"/etc/profiles/per-user/" + userName + "/share/applications",
	}

	// Also check XDG_DATA_DIRS environment variable
	if xdgDataDirs := os.Getenv("XDG_DATA_DIRS"); xdgDataDirs != "" {
		for _, dir := range strings.Split(xdgDataDirs, ":") {
			if dir != "" {
				paths = append(paths, filepath.Join(dir, "applications"))
			}
		}
	}

	return paths
}

// FilterExistingPaths returns only paths that exist on the filesystem
func FilterExistingPaths(paths []string) []string {
	var existing []string
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existing = append(existing, path)
		}
	}
	return existing
}
