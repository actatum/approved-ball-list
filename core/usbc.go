package core

import "context"

// USBC handles interfacing with the usbc api
type USBC interface {
	GetApprovedBallList(ctx context.Context) ([]Ball, error)
}
