package abl

import "context"

// Notification represents a type for notifying channels of a new ball.
type Notification struct {
	Ball Ball
}

// NotificationService represents a service for managing distribution of notifications.
type NotificationService interface {
	// SendNotifications sends notifications to the configured channels.
	SendNotifications(ctx context.Context, n ...Notification) error
}
