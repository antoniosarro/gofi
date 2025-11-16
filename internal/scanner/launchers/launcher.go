package launchers

import "github.com/antoniosarro/gofi/internal/domain/entry"

// GameLauncher represents a game launcher integration
type GameLauncher interface {
	// Name returns the launcher's identifier
	Name() string

	// Scan discovers and returns game entries from this launcher
	Scan() ([]*entry.Entry, error)
}
