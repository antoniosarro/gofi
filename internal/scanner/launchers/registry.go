package launchers

import (
	"sort"
	"sync"
)

var (
	registry = make(map[string]GameLauncher)
	mu       sync.RWMutex
)

// Register adds a game launcher to the registry
// This is typically called from init() functions in launcher implementations
func Register(launcher GameLauncher) {
	mu.Lock()
	defer mu.Unlock()
	registry[launcher.Name()] = launcher
}

// Get retrieves a launcher by name
func Get(name string) (GameLauncher, bool) {
	mu.RLock()
	defer mu.RUnlock()
	launcher, ok := registry[name]
	return launcher, ok
}

// GetAll returns all registered launchers
func GetAll() []GameLauncher {
	mu.RLock()
	defer mu.RUnlock()

	launchers := make([]GameLauncher, 0, len(registry))
	for _, launcher := range registry {
		launchers = append(launchers, launcher)
	}

	// Sort by name for consistent ordering
	sort.Slice(launchers, func(i, j int) bool {
		return launchers[i].Name() < launchers[j].Name()
	})

	return launchers
}

// List returns the names of all registered launchers
func List() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// Count returns the number of registered launchers
func Count() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(registry)
}

// Unregister removes a launcher from the registry (useful for testing)
func Unregister(name string) {
	mu.Lock()
	defer mu.Unlock()
	delete(registry, name)
}

// Clear removes all launchers from the registry (useful for testing)
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	registry = make(map[string]GameLauncher)
}
