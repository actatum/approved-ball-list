package core

import "context"

// Alerter handles sending alerts to the correct location
type Alerter interface {
	SendMessage(ctx context.Context, channelIDs []string, ball Ball) error
}
