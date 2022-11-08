package abl

import "context"

// USBCClient handles interfacing with USBC's api.
//
//go:generate moq -out ../mocks/usbc_client.go -pkg mocks -fmt goimports . USBCClient
type USBCClient interface {
	GetApprovedBallList(ctx context.Context) ([]Ball, error)
	Close()
}
