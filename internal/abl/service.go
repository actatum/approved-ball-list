// Package abl provides types and methods to handle business logic for the service.
package abl

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

// Service handles the business logic for the approved ball list service.
type Service interface {
	// RefreshBalls retrieves a list of balls from the USBC and a list of balls
	// from the service's repository, compares the two and stores any new ones in the
	// service's repository. It also deletes any ones that are in the repository but not
	// in the USBC's list. Then notifications are sent of the newly approved and revoked balls.
	RefreshBalls(ctx context.Context) error
}

type service struct {
	repo       Repository
	notifier   Notifier
	usbcClient USBCClient
}

// NewService returns a new instance of Service.
func NewService(repo Repository, notifier Notifier, usbcClient USBCClient) Service {
	return service{
		repo:       repo,
		notifier:   notifier,
		usbcClient: usbcClient,
	}
}

func (s service) RefreshBalls(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)

	usbcList, err := s.usbcClient.GetApprovedBallList(ctx)
	if err != nil {
		return fmt.Errorf("ubscClient.GetApprovedBallList: %w", err)
	}

	repoList, err := s.repo.ListBalls(ctx, BallFilter{})
	if err != nil {
		return fmt.Errorf("repo.ListBalls: %w", err)
	}

	approved := make([]Ball, 0)
	for _, ball := range usbcList {
		if !contains(repoList.Balls, ball) {
			logger.Info().Msgf("approved ball: %s %s", ball.Brand, ball.Name)
			approved = append(approved, ball)
		}
	}
	logger.Info().Msgf("%d newly approved balls", len(approved))

	if err := s.repo.AddBalls(ctx, approved); err != nil {
		return fmt.Errorf("repo.AddBalls: %w", err)
	}

	revoked := make([]Ball, 0)
	for _, ball := range repoList.Balls {
		if !contains(usbcList, ball) {
			logger.Info().Msgf("revoked ball: %s %s", ball.Brand, ball.Name)
			revoked = append(revoked, ball)
		}
	}
	logger.Info().Msgf("%d balls revoked", len(revoked))

	if err := s.repo.RemoveBalls(ctx, revoked); err != nil {
		return fmt.Errorf("repo.RemoveBalls: %w", err)
	}

	approvalNotifications := make([]Notification, 0, len(approved))
	for _, ball := range approved {
		approvalNotifications = append(approvalNotifications, Notification{
			Type: NotificationTypeApproved,
			Ball: ball,
		})
	}

	if err = s.notifier.Notify(ctx, approvalNotifications); err != nil {
		return fmt.Errorf("notifier.Notify: %w", err)
	}

	// revocationNotifications := make([]Notification, 0, len(revoked))
	// for _, ball := range revoked {
	// 	revocationNotifications = append(revocationNotifications, Notification{
	// 		Type: NotificationTypeRevoked,
	// 		Ball: ball,
	// 	})
	// }

	// if err = s.notifier.Notify(ctx, revocationNotifications); err != nil {
	// 	return fmt.Errorf("notifier.Notify: %w", err)
	// }

	return nil
}

func contains(s []Ball, b Ball) bool {
	for _, a := range s {
		if a.Equal(b) {
			return true
		}
	}
	return false
}
