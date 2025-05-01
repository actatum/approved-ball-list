// Package balls provides types for keeping up to date with the usbc approved ball list.
package balls

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

func Test_service_checkForNewlyApprovedBalls(t *testing.T) {
	t.Run("no newly approved balls", func(t *testing.T) {
		now := time.Now()

		hyroad := Ball{
			Brand:        Storm,
			Name:         "Hyroad",
			ApprovalDate: now,
		}

		usbcBalls := []Ball{hyroad}

		storeBalls := []Ball{hyroad}

		s := service{
			logger: slog.Default(),
			store: &StoreMock{
				GetAllBallsFunc: func(ctx context.Context, filter BallFilter) ([]Ball, error) {
					return storeBalls, nil
				},
			},
			usbcSerivce: &USBCServiceMock{
				ListBallsFunc: func(ctx context.Context, brand Brand) ([]Ball, error) {
					return usbcBalls, nil
				},
			},
		}

		jobs := make(chan Brand)
		results := make(chan jobResult)

		go s.checkForNewlyApprovedBalls(context.Background(), jobs, results)

		jobs <- Storm

		res := <-results

		close(jobs)
		close(results)

		if res.Err != nil {
			t.Fatal(res.Err)
		}

		if len(res.Balls) != 0 {
			t.Fatalf("expected 0 approved balls got %d", len(res.Balls))
		}
	})

	t.Run("one newly approved balls", func(t *testing.T) {
		now := time.Now()

		hyroad := Ball{
			Brand:        Storm,
			Name:         "Hyroad",
			ApprovalDate: now,
		}
		iqtour := Ball{
			Brand:        Storm,
			Name:         "!Q Tour",
			ApprovalDate: now,
		}

		usbcBalls := []Ball{hyroad, iqtour}

		storeBalls := []Ball{hyroad}

		s := service{
			logger: slog.Default(),
			store: &StoreMock{
				GetAllBallsFunc: func(ctx context.Context, filter BallFilter) ([]Ball, error) {
					return storeBalls, nil
				},
				AddBallsFunc: func(ctx context.Context, balls []Ball) error {
					return nil
				},
			},
			usbcSerivce: &USBCServiceMock{
				ListBallsFunc: func(ctx context.Context, brand Brand) ([]Ball, error) {
					return usbcBalls, nil
				},
			},
		}

		jobs := make(chan Brand)
		results := make(chan jobResult)

		go s.checkForNewlyApprovedBalls(context.Background(), jobs, results)

		jobs <- Storm

		res := <-results

		close(jobs)
		close(results)

		if res.Err != nil {
			t.Fatal(res.Err)
		}

		if len(res.Balls) != 1 {
			t.Fatalf("expected 1 approved balls got %d", len(res.Balls))
		}
	})

	t.Run("get all balls store error", func(t *testing.T) {
		now := time.Now()

		hyroad := Ball{
			Brand:        Storm,
			Name:         "Hyroad",
			ApprovalDate: now,
		}

		usbcBalls := []Ball{hyroad}

		s := service{
			logger: slog.Default(),
			store: &StoreMock{
				GetAllBallsFunc: func(ctx context.Context, filter BallFilter) ([]Ball, error) {
					return nil, fmt.Errorf("error")
				},
			},
			usbcSerivce: &USBCServiceMock{
				ListBallsFunc: func(ctx context.Context, brand Brand) ([]Ball, error) {
					return usbcBalls, nil
				},
			},
		}

		jobs := make(chan Brand)
		results := make(chan jobResult)

		go s.checkForNewlyApprovedBalls(context.Background(), jobs, results)

		jobs <- Storm

		res := <-results

		close(jobs)
		close(results)

		if res.Err == nil {
			t.Fatal("expected error got nil")
		}
	})

	t.Run("one newly approved balls store add error", func(t *testing.T) {
		now := time.Now()

		hyroad := Ball{
			Brand:        Storm,
			Name:         "Hyroad",
			ApprovalDate: now,
		}
		iqtour := Ball{
			Brand:        Storm,
			Name:         "!Q Tour",
			ApprovalDate: now,
		}

		usbcBalls := []Ball{hyroad, iqtour}

		storeBalls := []Ball{hyroad}

		s := service{
			logger: slog.Default(),
			store: &StoreMock{
				GetAllBallsFunc: func(ctx context.Context, filter BallFilter) ([]Ball, error) {
					return storeBalls, nil
				},
				AddBallsFunc: func(ctx context.Context, balls []Ball) error {
					return fmt.Errorf("error")
				},
			},
			usbcSerivce: &USBCServiceMock{
				ListBallsFunc: func(ctx context.Context, brand Brand) ([]Ball, error) {
					return usbcBalls, nil
				},
			},
		}

		jobs := make(chan Brand)
		results := make(chan jobResult)

		go s.checkForNewlyApprovedBalls(context.Background(), jobs, results)

		jobs <- Storm

		res := <-results

		close(jobs)
		close(results)

		if res.Err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
