package core

import "context"

// Repository handles interfacing with the persistence layer
type Repository interface {
	GetAllBalls(ctx context.Context) ([]Ball, error)
	InsertNewBalls(ctx context.Context, balls []Ball) error
}
