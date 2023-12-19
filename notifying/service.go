package notifying

import (
	"context"

	"github.com/actatum/approved-ball-list/notifications"
	"github.com/gofrs/uuid/v5"
)

type Service interface {
	RegisterTarget(ctx context.Context, targetType notifications.TargetType, destination string) (notifications.Target, error)
}

type service struct {
	notifications notifications.Repository
}

func NewService(notifications notifications.Repository) Service {
	return service{
		notifications: notifications,
	}
}

func (s service) RegisterTarget(ctx context.Context, targetType notifications.TargetType, destination string) (notifications.Target, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return notifications.Target{}, err
	}

	target := notifications.NewTarget(id.String(), targetType, destination)

	if err := s.notifications.StoreTarget(ctx, target); err != nil {
		return notifications.Target{}, err
	}

	return target, nil
}
