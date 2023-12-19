package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/actatum/approved-ball-list/balls"
)

type TargetType string

const (
	TargetTypeDiscord = "discord"
	TargetTypeEmail   = "email"
)

type State string

const (
	StatePending  = "pending"
	StateComplete = "complete"
	StateErrored  = "errored"
)

type Notification struct {
	ID      string
	State   State
	Content []balls.Ball
}

type Target struct {
	ID   string
	Type TargetType
	// The email account or discord channel to send the notification to.
	Destination string

	CreatedAt time.Time
	UpdateAt  time.Time
}

func NewTarget(id string, targetType TargetType, destination string) Target {
	now := time.Now()
	return Target{
		ID:          id,
		Type:        targetType,
		Destination: destination,
		CreatedAt:   now,
		UpdateAt:    now,
	}
}

type Repository interface {
	StoreTarget(ctx context.Context, target Target) error
	FindAll(ctx context.Context) ([]Target, error)
}

type DuplicateTargetError struct {
	targetType  TargetType
	destination string
}

func (e DuplicateTargetError) Error() string {
	return fmt.Sprintf("target already exists for %s: %s", e.targetType, e.destination)
}
