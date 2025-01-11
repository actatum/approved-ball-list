// Package balls provides types for keeping up to date with the usbc approved ball list.
package balls

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/rs/zerolog"
)

//go:generate moq -out ../mocks/notification_service.go -pkg mocks -fmt goimports . NotificationService
//go:generate moq -out ../mocks/repository.go -pkg mocks -fmt goimports . Repository
//go:generate moq -out ../mocks/usbc_service.go -pkg mocks -fmt goimports . USBCService

// Brand is a brand that makes bowling equipment.
type Brand string

// All active brands.
const (
	Global      Brand = "900 Global"
	BigBowling  Brand = "BIG Bowling"
	Brunswick   Brand = "Brunswick"
	Columbia300 Brand = "Columbia"
	DV8         Brand = "DV8"
	Ebonite     Brand = "Ebonite"
	Hammer      Brand = "Hammer"
	Motiv       Brand = "Motiv"
	Radical     Brand = "Radical"
	RotoGrip    Brand = "Roto Grip"
	Storm       Brand = "Storm"
	Swag        Brand = "Swag"
	Track       Brand = "Track Inc."
)

var allBrands = []Brand{
	Global,
	BigBowling,
	Brunswick,
	Columbia300,
	DV8,
	Ebonite,
	Hammer,
	Motiv,
	Radical,
	RotoGrip,
	Storm,
	Swag,
	Track,
}

// Ball represents a bowling ball.
type Ball struct {
	ID           int
	Brand        Brand
	Name         string
	ApprovalDate time.Time
	ImageURL     *url.URL
}

// USBCService interacts with the USBC approved ball list.
type USBCService interface {
	ListBalls(ctx context.Context, brand Brand) ([]Ball, error)
}

// NotificationService handles sending notifications.
type NotificationService interface {
	SendNotification(ctx context.Context, approvedBalls []Ball) error
}

// Repository interacts with the persistence for the service.
type Repository interface {
	Add(ctx context.Context, balls ...Ball) ([]Ball, error)
}

// Service handles business logic.
type Service interface {
	// CheckForNewlyApprovedBalls checks to see if any new balls are on the USBC approved ball list.
	CheckForNewlyApprovedBalls(ctx context.Context) error
}

type service struct {
	repo                Repository
	usbcSerivce         USBCService
	notificationService NotificationService
}

// NewService returns a new service.
func NewService(
	repo Repository,
	ubscService USBCService,
	notificationService NotificationService,
) Service {
	return service{
		repo:                repo,
		usbcSerivce:         ubscService,
		notificationService: notificationService,
	}
}

type jobResult struct {
	Balls []Ball
	Err   error
}

func (s service) CheckForNewlyApprovedBalls(ctx context.Context) error {
	numJobs := len(allBrands)
	jobs := make(chan Brand, numJobs)
	results := make(chan jobResult, numJobs)

	for w := 0; w < 6; w++ {
		go s.checkForNewlyApprovedBalls(ctx, jobs, results)
	}

	for j := 0; j < numJobs; j++ {
		jobs <- allBrands[j]
	}
	close(jobs)

	approved := make([]Ball, 0)
	var err error
	for r := 0; r < numJobs; r++ {
		res := <-results
		if res.Err != nil {
			err = errors.Join(err, res.Err)
		}
		if len(res.Balls) > 0 {
			approved = append(approved, res.Balls...)
		}
	}
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("error checking for approved balls")
	}

	zerolog.Ctx(ctx).Info().Msgf("%d newly approved balls", len(approved))

	if err := s.notificationService.SendNotification(ctx, approved); err != nil {
		return fmt.Errorf("sending notification: %w", err)
	}

	return nil
}

func (s service) checkForNewlyApprovedBalls(ctx context.Context, jobs <-chan Brand, results chan<- jobResult) {
	for j := range jobs {
		zerolog.Ctx(ctx).Info().Msgf("listing balls from %s", j)
		balls, err := s.usbcSerivce.ListBalls(ctx, j)
		if err != nil {
			results <- jobResult{
				Err: fmt.Errorf("checking usbc list: %w", err),
			}
		}

		if len(balls) == 0 {
			continue
		}

		added, err := s.repo.Add(ctx, balls...)
		if err != nil {
			results <- jobResult{
				Err: fmt.Errorf("adding balls to repo: %w", err),
			}
		}

		results <- jobResult{
			Balls: added,
		}
	}
}
