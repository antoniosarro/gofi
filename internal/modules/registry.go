package modules

import (
	"fmt"
	"sort"
	"sync"
)

var (
	// registry holds all registered modules
	registry = make(map[string]Module)

	// mu protects concurrent access to registry
	mu sync.RWMutex
)

// Register adds a module to the registry.
// This is typically called from init() functions in module implementations.
// Registering a module with the same name twice will overwrite the previous registration.
//
// Example:
//
//	func init() {
//	    modules.Register(&MyModule{})
//	}
func Register(m Module) {
	mu.Lock()
	defer mu.Unlock()
	registry[m.Name()] = m
}

// Get retrieves a module by name.
// Returns an error if the module is not found.
//
// Example:
//
//	module, err := modules.Get("application")
//	if err != nil {
//	    log.Fatal(err)
//	}
func Get(name string) (Module, error) {
	mu.RLock()
	defer mu.RUnlock()

	m, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("module '%s' not found", name)
	}
	return m, nil
}

// List returns all registered module names in alphabetical order.
// Useful for displaying available modules to users.
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

// All returns all registered modules as a map.
// The returned map is a copy and can be safely modified.
func All() map[string]Module {
	mu.RLock()
	defer mu.RUnlock()

	result := make(map[string]Module, len(registry))
	for name, module := range registry {
		result[name] = module
	}
	return result
}

// Count returns the number of registered modules.
func Count() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(registry)
}

// Unregister removes a module from the registry.
// This is primarily useful for testing.
func Unregister(name string) {
	mu.Lock()
	defer mu.Unlock()
	delete(registry, name)
}

// Clear removes all modules from the registry.
// This is primarily useful for testing.
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	registry = make(map[string]Module)
}
