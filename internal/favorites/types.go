package favorites

import "time"

const (
	// EventWeightLaunch determines importance of launch events
	EventWeightLaunch = 1.0

	// EventWeightSearch determines importance of search events
	EventWeightSearch = 0.3

	// DecayLambda controls how fast old events lose importance
	// 0.1 = noticeable decay after ~7 days
	DecayLambda = 0.1

	// FavoriteThreshold is the minimum score to be considered a favorite
	FavoriteThreshold = 5.0

	// MaxEvents limits memory usage
	MaxEvents = 1000
)

// EventType represents the type of user interaction
type EventType string

const (
	EventTypeLaunch EventType = "launch"
	EventTypeSearch EventType = "search"
)

// Event represents a user interaction with an app
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Type      EventType `json:"type"`
}

// AppStats tracks usage statistics for an application
type AppStats struct {
	DesktopFile string  `json:"desktop_file"` // Unique identifier
	Events      []Event `json:"events"`
	Score       float64 `json:"-"` // Computed at runtime
}
