package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSimple(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	content := `name = "test"
version = 1
enabled = true
`

	os.WriteFile(configPath, []byte(content), 0644)

	parser := New()
	data, err := parser.ParseFile(configPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if name, ok := data.GetString("name"); !ok || name != "test" {
		t.Errorf("name = %q, want %q", name, "test")
	}

	if version, ok := data.GetInt("version"); !ok || version != 1 {
		t.Errorf("version = %d, want 1", version)
	}

	if enabled, ok := data.GetBool("enabled"); !ok || !enabled {
		t.Error("enabled should be true")
	}
}

func TestParseTable(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	content := `[section]
key = "value"
number = 42
`

	os.WriteFile(configPath, []byte(content), 0644)

	parser := New()
	data, err := parser.ParseFile(configPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	section, ok := data.GetTable("section")
	if !ok {
		t.Fatal("section table not found")
	}

	if key, ok := section.GetString("key"); !ok || key != "value" {
		t.Errorf("key = %q, want %q", key, "value")
	}

	if number, ok := section.GetInt("number"); !ok || number != 42 {
		t.Errorf("number = %d, want 42", number)
	}
}

func TestParseNestedTable(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	content := `[parent.child]
value = "nested"
`

	os.WriteFile(configPath, []byte(content), 0644)

	parser := New()
	data, err := parser.ParseFile(configPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	parent, ok := data.GetTable("parent")
	if !ok {
		t.Fatal("parent table not found")
	}

	child, ok := parent.GetTable("child")
	if !ok {
		t.Fatal("child table not found")
	}

	if value, ok := child.GetString("value"); !ok || value != "nested" {
		t.Errorf("value = %q, want %q", value, "nested")
	}
}

func TestParseArray(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	content := `items = ["one", "two", "three"]
numbers = [1, 2, 3]
`

	os.WriteFile(configPath, []byte(content), 0644)

	parser := New()
	data, err := parser.ParseFile(configPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	items, ok := data.GetStringSlice("items")
	if !ok {
		t.Fatal("items array not found")
	}

	expected := []string{"one", "two", "three"}
	if len(items) != len(expected) {
		t.Errorf("items length = %d, want %d", len(items), len(expected))
	}

	for i, item := range items {
		if item != expected[i] {
			t.Errorf("items[%d] = %q, want %q", i, item, expected[i])
		}
	}
}

func TestParseComments(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	content := `# This is a comment
name = "test"
# Another comment
version = 1
`

	os.WriteFile(configPath, []byte(content), 0644)

	parser := New()
	data, err := parser.ParseFile(configPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if name, ok := data.GetString("name"); !ok || name != "test" {
		t.Errorf("name = %q, want %q", name, "test")
	}
}

func TestParseBooleans(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	content := `enabled = true
disabled = false
`

	os.WriteFile(configPath, []byte(content), 0644)

	parser := New()
	data, err := parser.ParseFile(configPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if enabled, ok := data.GetBool("enabled"); !ok || !enabled {
		t.Error("enabled should be true")
	}

	if disabled, ok := data.GetBool("disabled"); !ok || disabled {
		t.Error("disabled should be false")
	}
}
