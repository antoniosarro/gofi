package entry

import (
	"os"
	"testing"
)

func TestGetAppType(t *testing.T) {
	tests := []struct {
		name     string
		entry    *Entry
		expected AppType
	}{
		{
			name: "Flatpak application",
			entry: &Entry{
				Name: "Firefox",
				Path: "/var/lib/flatpak/exports/share/applications/org.mozilla.firefox.desktop",
			},
			expected: AppTypeFlatpak,
		},
		{
			name: "NixOS system application",
			entry: &Entry{
				Name: "Firefox",
				Path: "/run/current-system/sw/share/applications/firefox.desktop",
			},
			expected: AppTypeNixSystem,
		},
		{
			name: "NixOS home-manager application",
			entry: &Entry{
				Name: "Alacritty",
				Path: os.Getenv("HOME") + "/.nix-profile/share/applications/alacritty.desktop",
			},
			expected: AppTypeNixHome,
		},
		{
			name: "System application",
			entry: &Entry{
				Name: "Firefox",
				Path: "/usr/share/applications/firefox.desktop",
			},
			expected: AppTypeSystem,
		},
		{
			name: "Game application",
			entry: &Entry{
				Name:       "SuperTuxKart",
				Path:       "/usr/share/applications/supertuxkart.desktop",
				Categories: []string{"Game", "ArcadeGame"},
			},
			expected: AppTypeGame,
		},
		{
			name: "Other application",
			entry: &Entry{
				Name: "CustomApp",
				Path: "/home/user/.local/share/applications/custom.desktop",
			},
			expected: AppTypeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entry.GetAppType()
			if result != tt.expected {
				t.Errorf("GetAppType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		entry   *Entry
		wantErr error
	}{
		{
			name: "Valid entry",
			entry: &Entry{
				Name: "Firefox",
				Exec: "firefox",
			},
			wantErr: nil,
		},
		{
			name: "Missing name",
			entry: &Entry{
				Exec: "firefox",
			},
			wantErr: ErrMissingName,
		},
		{
			name: "Missing exec",
			entry: &Entry{
				Name: "Firefox",
			},
			wantErr: ErrMissingExec,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClone(t *testing.T) {
	original := &Entry{
		Name:        "Firefox",
		GenericName: "Web Browser",
		Comment:     "Browse the Web",
		Exec:        "firefox",
		Icon:        "firefox",
		Terminal:    false,
		Categories:  []string{"Network", "WebBrowser"},
		Path:        "/usr/share/applications/firefox.desktop",
	}

	cloned := original.Clone()

	// Check if values are equal
	if cloned.Name != original.Name {
		t.Errorf("Clone() Name = %v, want %v", cloned.Name, original.Name)
	}

	// Modify cloned categories to ensure deep copy
	cloned.Categories[0] = "Modified"
	if original.Categories[0] == "Modified" {
		t.Error("Clone() did not create a deep copy of Categories")
	}
}

func TestParseExecString(t *testing.T) {
	tests := []struct {
		name     string
		exec     string
		expected string
	}{
		{
			name:     "Simple command",
			exec:     "firefox",
			expected: "firefox",
		},
		{
			name:     "Command with field codes",
			exec:     "firefox %u",
			expected: "firefox",
		},
		{
			name:     "Command with multiple field codes",
			exec:     "gimp %f %F",
			expected: "gimp",
		},
		{
			name:     "Command with extra spaces",
			exec:     "firefox  %u   %U",
			expected: "firefox",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseExecString(tt.exec)
			if result != tt.expected {
				t.Errorf("parseExecString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSplitExecString(t *testing.T) {
	tests := []struct {
		name     string
		exec     string
		expected []string
	}{
		{
			name:     "Simple command",
			exec:     "firefox",
			expected: []string{"firefox"},
		},
		{
			name:     "Command with arguments",
			exec:     "firefox --new-window",
			expected: []string{"firefox", "--new-window"},
		},
		{
			name:     "Command with quoted argument",
			exec:     `firefox "https://example.com"`,
			expected: []string{"firefox", "https://example.com"},
		},
		{
			name:     "Command with single quotes",
			exec:     `echo 'hello world'`,
			expected: []string{"echo", "hello world"},
		},
		{
			name:     "Command with escaped characters",
			exec:     `echo hello\ world`,
			expected: []string{"echo", "hello world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitExecString(tt.exec)
			if len(result) != len(tt.expected) {
				t.Errorf("splitExecString() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitExecString()[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestIsValidEnvVarName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid uppercase", "PATH", true},
		{"Valid with underscore", "MY_VAR", true},
		{"Valid with numbers", "VAR123", true},
		{"Invalid starts with number", "123VAR", false},
		{"Invalid with space", "MY VAR", false},
		{"Invalid with dash", "MY-VAR", false},
		{"Empty string", "", false},
		{"Valid lowercase", "path", true},
		{"Valid mixed case", "MyVar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEnvVarName(tt.input)
			if result != tt.expected {
				t.Errorf("isValidEnvVarName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
