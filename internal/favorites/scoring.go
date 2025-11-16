package favorites

import (
	"math"
	"time"
)

// Scorer handles score calculations for favorites
type Scorer struct{}

// NewScorer creates a new scorer
func NewScorer() *Scorer {
	return &Scorer{}
}

// CalculateScore computes the time-decay weighted score
func (s *Scorer) CalculateScore(stats *AppStats) float64 {
	if stats == nil || len(stats.Events) == 0 {
		return 0
	}

	now := time.Now()
	score := 0.0

	for _, event := range stats.Events {
		// Calculate age in days
		age := now.Sub(event.Timestamp).Hours() / 24.0

		// Get event weight
		weight := s.getEventWeight(event.Type)

		// Apply exponential time decay: weight * e^(-λ * age)
		score += weight * math.Exp(-DecayLambda*age)
	}

	return score
}

// getEventWeight returns the weight for an event type
func (s *Scorer) getEventWeight(eventType EventType) float64 {
	switch eventType {
	case EventTypeLaunch:
		return EventWeightLaunch
	case EventTypeSearch:
		return EventWeightSearch
	default:
		return 0
	}
}

// IsFavorite checks if a score qualifies as a favorite
func (s *Scorer) IsFavorite(score float64) bool {
	return score >= FavoriteThreshold
}
