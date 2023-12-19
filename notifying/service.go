package notifying

import (
	"context"

	"github.com/actatum/approved-ball-list/balls"
	"github.com/actatum/approved-ball-list/notifications"
	"github.com/gofrs/uuid/v5"
)

type Service interface {
	RegisterTarget(ctx context.Context, targetType notifications.TargetType, destination string) (notifications.Target, error)
	CreateNotifications(ctx context.Context, approvedBall balls.Ball) error
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

func (s service) CreateNotifications(ctx context.Context, approvedBall balls.Ball) error {
	targets, err := s.notifications.FindAllTargets(ctx)
	if err != nil {
		return err
	}

	notifs := make([]notifications.Notification, 0, len(targets))
	for _, target := range targets {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}

		notif := notifications.NewNotification(id.String(), []balls.Ball{approvedBall}, target)
		notifs = append(notifs, notif)
	}

	if err := s.notifications.Store(ctx, notifs); err != nil {
		return err
	}

	return nil
}
