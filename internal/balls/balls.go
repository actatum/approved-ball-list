// Package balls provides types for keeping up to date with the usbc approved ball list.
package balls

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"
)

type Service interface {
	// CheckForNewlyApprovedBalls checks to see if any new balls are on the USBC approved ball list.
	CheckForNewlyApprovedBalls(ctx context.Context) error
}

// Ball represents a bowling ball.
type Ball struct {
	ID           int
	Brand        Brand
	Name         string
	ApprovalDate time.Time
	ImageURL     *url.URL
}

func BallsEqual(b1 Ball, b2 Ball) bool {
	if b1.Brand != b2.Brand {
		return false
	}

	if b1.Name != b2.Name {
		return false
	}

	if !b1.ApprovalDate.Equal(b2.ApprovalDate) {
		return false
	}

	return true
}

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

type BallFilter struct {
	Brand        *Brand
	Name         *string
	ApprovalDate *time.Time
}

type service struct {
	logger      *slog.Logger
	store       Store
	usbcSerivce USBCService
	notifier    Notifier
}

// NewService returns a new service.
func NewService(
	logger *slog.Logger,
	store Store,
	ubscService USBCService,
	notifier Notifier,
) Service {
	return service{
		logger:      logger,
		store:       store,
		usbcSerivce: ubscService,
		notifier:    notifier,
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

	for w := 0; w < 7; w++ {
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
		s.logger.ErrorContext(ctx, "error checking for approved balls", slog.Any("error", err))
	}

	s.logger.InfoContext(ctx, fmt.Sprintf("%d newly approved balls", len(approved)))

	if err := s.notifier.Notify(ctx, approved); err != nil {
		return fmt.Errorf("notifying: %w", err)
	}

	return nil
}

func (s service) checkForNewlyApprovedBalls(ctx context.Context, jobs <-chan Brand, results chan<- jobResult) {
	for brand := range jobs {
		s.logger.InfoContext(ctx, fmt.Sprintf("listing balls from %s", brand))
		balls, err := s.usbcSerivce.ListBalls(ctx, brand)
		if err != nil {
			results <- jobResult{
				Err: fmt.Errorf("checking usbc list for brand %s: %w", brand, err),
			}
			continue
		}

		if len(balls) == 0 {
			continue
		}

		brandBalls, err := s.store.GetAllBalls(ctx, BallFilter{
			Brand: &brand,
		})
		if err != nil {
			results <- jobResult{
				Err: fmt.Errorf("retrieving balls for brand %s from store: %w", brand, err),
			}
			continue
		}

		approved := make([]Ball, 0)
		for _, usbcBall := range balls {
			found := false
			for _, storedBall := range brandBalls {
				if BallsEqual(usbcBall, storedBall) {
					found = true
					break
				}
			}
			if !found {
				approved = append(approved, usbcBall)
			}
		}

		if len(approved) == 0 {
			results <- jobResult{
				Balls: approved,
			}
			return
		}

		if err = s.store.AddBalls(ctx, approved); err != nil {
			results <- jobResult{
				Err: fmt.Errorf("adding balls to store: %w", err),
			}
			continue
		}

		results <- jobResult{
			Balls: approved,
		}
	}
}
