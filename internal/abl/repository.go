package abl

import "context"

// Repository handles interfacing with the service's persistence layer.
//
//go:generate moq -out ../mocks/repository.go -pkg mocks -fmt goimports . Repository
type Repository interface {
	AddBalls(ctx context.Context, balls []Ball) error
	ListBalls(ctx context.Context, filter BallFilter) (ListBallsResult, error)
	RemoveBalls(ctx context.Context, balls []Ball) error
}
