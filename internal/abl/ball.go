package abl

import (
	"context"
	"time"
)

// ActiveBrands is a map of all active brands.
var ActiveBrands = map[string]bool{
	"900 Global":  true,
	"BIG Bowling": true,
	"Brunswick":   true,
	"Columbia":    true,
	"DV8":         true,
	"Ebonite":     true,
	"Hammer":      true,
	"Motiv":       true,
	"Radical":     true,
	"Roto Grip":   true,
	"Storm":       true,
	"Swag":        true,
	"Track Inc.":  true,
}

// Ball represents a bowling ball.
type Ball struct {
	ID         int       `json:"id"`
	Brand      string    `json:"brand"`
	Name       string    `json:"name"`
	ApprovedAt time.Time `json:"approved_at"`
	ImageURL   string    `json:"image_url"`
}

// Equal is used to determine if two balls are the same.
func (b Ball) Equal(b2 Ball) bool {
	return b.Name == b2.Name && b.Brand == b2.Brand
}

// BallService represents a service for managing our copy of the USBC's approved ball list.
type BallService interface {
	// AddBalls adds a batch of balls.
	AddBalls(ctx context.Context, balls ...Ball) error
	// ListBalls returns a list of all balls with the given filter.
	ListBalls(ctx context.Context, filter BallFilter) ([]Ball, int, error)
	// RemoveBalls removes a batch of balls.
	RemoveBalls(ctx context.Context, balls ...Ball) error
}

// BallFilter is used to filter the returns from ListBalls.
type BallFilter struct {
	Brand *string

	PageSize  int
	PageToken string
}

// ListBallsResult collects results for listing balls.
type ListBallsResult struct {
	Balls         []Ball
	NextPageToken string
	Count         int
}
