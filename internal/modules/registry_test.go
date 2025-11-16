package modules

import (
	"testing"

	"github.com/antoniosarro/gofi/internal/config"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// MockModule is a test module implementation
type MockModule struct {
	name        string
	description string
	initialized bool
	cleaned     bool
}

func (m *MockModule) Name() string                                      { return m.name }
func (m *MockModule) Description() string                               { return m.description }
func (m *MockModule) Initialize(cfg *config.ModuleConfig) error         { m.initialized = true; return nil }
func (m *MockModule) CreateWindow(app *gtk.Application) (Window, error) { return nil, nil }
func (m *MockModule) Cleanup() error                                    { m.cleaned = true; return nil }

func TestRegister(t *testing.T) {
	// Clear registry for test
	Clear()

	module := &MockModule{name: "test", description: "Test module"}
	Register(module)

	if Count() != 1 {
		t.Errorf("Count() = %d, want 1", Count())
	}

	retrieved, err := Get("test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.Name() != "test" {
		t.Errorf("Retrieved module name = %s, want test", retrieved.Name())
	}
}

func TestGet(t *testing.T) {
	Clear()

	module := &MockModule{name: "test", description: "Test module"}
	Register(module)

	tests := []struct {
		name       string
		moduleName string
		wantErr    bool
	}{
		{"Existing module", "test", false},
		{"Non-existing module", "nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Get(tt.moduleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestList(t *testing.T) {
	Clear()

	modules := []*MockModule{
		{name: "zebra", description: "Zebra module"},
		{name: "alpha", description: "Alpha module"},
		{name: "beta", description: "Beta module"},
	}

	for _, m := range modules {
		Register(m)
	}

	list := List()

	// Should be sorted alphabetically
	expected := []string{"alpha", "beta", "zebra"}
	if len(list) != len(expected) {
		t.Fatalf("List() length = %d, want %d", len(list), len(expected))
	}

	for i, name := range list {
		if name != expected[i] {
			t.Errorf("List()[%d] = %s, want %s", i, name, expected[i])
		}
	}
}

func TestAll(t *testing.T) {
	Clear()

	module1 := &MockModule{name: "test1", description: "Test 1"}
	module2 := &MockModule{name: "test2", description: "Test 2"}

	Register(module1)
	Register(module2)

	all := All()

	if len(all) != 2 {
		t.Errorf("All() length = %d, want 2", len(all))
	}

	if _, ok := all["test1"]; !ok {
		t.Error("All() missing test1")
	}

	if _, ok := all["test2"]; !ok {
		t.Error("All() missing test2")
	}
}

func TestCount(t *testing.T) {
	Clear()

	if Count() != 0 {
		t.Errorf("Initial Count() = %d, want 0", Count())
	}

	Register(&MockModule{name: "test1", description: "Test 1"})
	if Count() != 1 {
		t.Errorf("Count() after one registration = %d, want 1", Count())
	}

	Register(&MockModule{name: "test2", description: "Test 2"})
	if Count() != 2 {
		t.Errorf("Count() after two registrations = %d, want 2", Count())
	}
}

func TestUnregister(t *testing.T) {
	Clear()

	Register(&MockModule{name: "test", description: "Test"})

	if Count() != 1 {
		t.Fatalf("Setup failed: Count() = %d, want 1", Count())
	}

	Unregister("test")

	if Count() != 0 {
		t.Errorf("Count() after Unregister = %d, want 0", Count())
	}

	_, err := Get("test")
	if err == nil {
		t.Error("Get() after Unregister should return error")
	}
}

func TestClear(t *testing.T) {
	Clear()

	Register(&MockModule{name: "test1", description: "Test 1"})
	Register(&MockModule{name: "test2", description: "Test 2"})
	Register(&MockModule{name: "test3", description: "Test 3"})

	if Count() != 3 {
		t.Fatalf("Setup failed: Count() = %d, want 3", Count())
	}

	Clear()

	if Count() != 0 {
		t.Errorf("Count() after Clear = %d, want 0", Count())
	}
}

func TestRegisterOverwrite(t *testing.T) {
	Clear()

	module1 := &MockModule{name: "test", description: "First"}
	module2 := &MockModule{name: "test", description: "Second"}

	Register(module1)
	Register(module2)

	if Count() != 1 {
		t.Errorf("Count() = %d, want 1 (should overwrite)", Count())
	}

	retrieved, _ := Get("test")
	if retrieved.Description() != "Second" {
		t.Errorf("Description = %s, want Second", retrieved.Description())
	}
}

func TestConcurrentAccess(t *testing.T) {
	Clear()

	// Test concurrent registration and retrieval
	done := make(chan bool, 3)

	// Goroutine 1: Register modules
	go func() {
		for i := 0; i < 100; i++ {
			Register(&MockModule{name: "concurrent", description: "Test"})
		}
		done <- true
	}()

	// Goroutine 2: Get modules
	go func() {
		for i := 0; i < 100; i++ {
			Get("concurrent")
		}
		done <- true
	}()

	// Goroutine 3: List modules
	go func() {
		for i := 0; i < 100; i++ {
			List()
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}
