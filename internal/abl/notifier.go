package abl

import "context"

// NotificationType represents the enum type for different notifications.
type NotificationType string

// All possible values for notification type.
const (
	NotificationTypeApproved = "approved"
	NotificationTypeRevoked  = "revoked"
)

// Notification represents a type for notifying channels of a new ball.
type Notification struct {
	Type NotificationType
	Ball Ball
}

// Notifier manages sending notifications for the service.
//
//go:generate moq -out ../mocks/notifier.go -pkg mocks -fmt goimports . Notifier
type Notifier interface {
	// Notify sends notifications.
	Notify(ctx context.Context, notifications []Notification) error
	Close() error
}
