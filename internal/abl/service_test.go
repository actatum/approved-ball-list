package abl_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/actatum/approved-ball-list/internal/abl"
	"github.com/actatum/approved-ball-list/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_service_RefreshBalls(t *testing.T) {
	type expectations struct {
		repoAddCalls    int
		repoListCalls   int
		repoRemoveCalls int

		notifierNotifyCalls int

		usbcClientGetCalls int
	}
	type fields struct {
		repo       abl.Repository
		notifier   abl.Notifier
		usbcClient abl.USBCClient
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    expectations
	}{
		{
			name: "success with no bowling balls in repo or usbc list",
			fields: fields{
				repo: &mocks.RepositoryMock{
					AddBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return nil
					},
					ListBallsFunc: func(ctx context.Context, filter abl.BallFilter) (abl.ListBallsResult, error) {
						return abl.ListBallsResult{}, nil
					},
					RemoveBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return nil
					},
				},
				notifier: &mocks.NotifierMock{
					NotifyFunc: func(ctx context.Context, notifications []abl.Notification) error {
						return nil
					},
				},
				usbcClient: &mocks.USBCClientMock{
					GetApprovedBallListFunc: func(ctx context.Context) ([]abl.Ball, error) {
						return nil, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
			want: expectations{
				repoAddCalls:        1,
				repoListCalls:       1,
				repoRemoveCalls:     1,
				notifierNotifyCalls: 1,
				usbcClientGetCalls:  1,
			},
		},
		{
			name: "success with some bowling balls in repo and none in usbc list",
			fields: fields{
				repo: &mocks.RepositoryMock{
					AddBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return nil
					},
					ListBallsFunc: func(ctx context.Context, filter abl.BallFilter) (abl.ListBallsResult, error) {
						return abl.ListBallsResult{
							Balls: []abl.Ball{
								{
									ID:         1,
									Brand:      "Storm",
									Name:       "Super Nova",
									ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
									ImageURL:   "image",
								},
								{
									ID:         2,
									Brand:      "Storm",
									Name:       "Phaze V",
									ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
									ImageURL:   "image",
								},
							},
							Count: 2,
						}, nil
					},
					RemoveBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return nil
					},
				},
				notifier: &mocks.NotifierMock{
					NotifyFunc: func(ctx context.Context, notifications []abl.Notification) error {
						return nil
					},
				},
				usbcClient: &mocks.USBCClientMock{
					GetApprovedBallListFunc: func(ctx context.Context) ([]abl.Ball, error) {
						return nil, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
			want: expectations{
				repoAddCalls:        1,
				repoListCalls:       1,
				repoRemoveCalls:     1,
				notifierNotifyCalls: 1,
				usbcClientGetCalls:  1,
			},
		},
		{
			name: "success with no bowling balls in repo but some in usbc list",
			fields: fields{
				repo: &mocks.RepositoryMock{
					AddBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return nil
					},
					ListBallsFunc: func(ctx context.Context, filter abl.BallFilter) (abl.ListBallsResult, error) {
						return abl.ListBallsResult{}, nil
					},
					RemoveBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return nil
					},
				},
				notifier: &mocks.NotifierMock{
					NotifyFunc: func(ctx context.Context, notifications []abl.Notification) error {
						return nil
					},
				},
				usbcClient: &mocks.USBCClientMock{
					GetApprovedBallListFunc: func(ctx context.Context) ([]abl.Ball, error) {
						return []abl.Ball{
							{
								ID:         1,
								Brand:      "Storm",
								Name:       "Super Nova",
								ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
								ImageURL:   "image",
							},
							{
								ID:         2,
								Brand:      "Storm",
								Name:       "Phaze V",
								ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
								ImageURL:   "image",
							},
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
			want: expectations{
				repoAddCalls:        1,
				repoListCalls:       1,
				repoRemoveCalls:     1,
				notifierNotifyCalls: 1,
				usbcClientGetCalls:  1,
			},
		},
		{
			name: "add balls error",
			fields: fields{
				repo: &mocks.RepositoryMock{
					AddBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return fmt.Errorf("add balls error")
					},
					ListBallsFunc: func(ctx context.Context, filter abl.BallFilter) (abl.ListBallsResult, error) {
						return abl.ListBallsResult{}, nil
					},
					RemoveBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return nil
					},
				},
				notifier: &mocks.NotifierMock{
					NotifyFunc: func(ctx context.Context, notifications []abl.Notification) error {
						return nil
					},
				},
				usbcClient: &mocks.USBCClientMock{
					GetApprovedBallListFunc: func(ctx context.Context) ([]abl.Ball, error) {
						return []abl.Ball{
							{
								ID:         1,
								Brand:      "Storm",
								Name:       "Super Nova",
								ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
								ImageURL:   "image",
							},
							{
								ID:         2,
								Brand:      "Storm",
								Name:       "Phaze V",
								ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
								ImageURL:   "image",
							},
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
			want: expectations{
				repoAddCalls:        1,
				repoListCalls:       1,
				repoRemoveCalls:     0,
				notifierNotifyCalls: 0,
				usbcClientGetCalls:  1,
			},
		},
		{
			name: "remove balls error",
			fields: fields{
				repo: &mocks.RepositoryMock{
					AddBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return nil
					},
					ListBallsFunc: func(ctx context.Context, filter abl.BallFilter) (abl.ListBallsResult, error) {
						return abl.ListBallsResult{}, nil
					},
					RemoveBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return fmt.Errorf("remove balls error")
					},
				},
				notifier: &mocks.NotifierMock{
					NotifyFunc: func(ctx context.Context, notifications []abl.Notification) error {
						return nil
					},
				},
				usbcClient: &mocks.USBCClientMock{
					GetApprovedBallListFunc: func(ctx context.Context) ([]abl.Ball, error) {
						return []abl.Ball{
							{
								ID:         1,
								Brand:      "Storm",
								Name:       "Super Nova",
								ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
								ImageURL:   "image",
							},
							{
								ID:         2,
								Brand:      "Storm",
								Name:       "Phaze V",
								ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
								ImageURL:   "image",
							},
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
			want: expectations{
				repoAddCalls:        1,
				repoListCalls:       1,
				repoRemoveCalls:     1,
				notifierNotifyCalls: 0,
				usbcClientGetCalls:  1,
			},
		},
		{
			name: "notify error",
			fields: fields{
				repo: &mocks.RepositoryMock{
					AddBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return nil
					},
					ListBallsFunc: func(ctx context.Context, filter abl.BallFilter) (abl.ListBallsResult, error) {
						return abl.ListBallsResult{}, nil
					},
					RemoveBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return nil
					},
				},
				notifier: &mocks.NotifierMock{
					NotifyFunc: func(ctx context.Context, notifications []abl.Notification) error {
						return fmt.Errorf("notify error")
					},
				},
				usbcClient: &mocks.USBCClientMock{
					GetApprovedBallListFunc: func(ctx context.Context) ([]abl.Ball, error) {
						return []abl.Ball{
							{
								ID:         1,
								Brand:      "Storm",
								Name:       "Super Nova",
								ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
								ImageURL:   "image",
							},
							{
								ID:         2,
								Brand:      "Storm",
								Name:       "Phaze V",
								ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
								ImageURL:   "image",
							},
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
			want: expectations{
				repoAddCalls:        1,
				repoListCalls:       1,
				repoRemoveCalls:     1,
				notifierNotifyCalls: 1,
				usbcClientGetCalls:  1,
			},
		},
		{
			name: "get approved ball list error",
			fields: fields{
				repo: &mocks.RepositoryMock{
					AddBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return fmt.Errorf("add balls error")
					},
					ListBallsFunc: func(ctx context.Context, filter abl.BallFilter) (abl.ListBallsResult, error) {
						return abl.ListBallsResult{}, nil
					},
					RemoveBallsFunc: func(ctx context.Context, balls []abl.Ball) error {
						return nil
					},
				},
				notifier: &mocks.NotifierMock{
					NotifyFunc: func(ctx context.Context, notifications []abl.Notification) error {
						return nil
					},
				},
				usbcClient: &mocks.USBCClientMock{
					GetApprovedBallListFunc: func(ctx context.Context) ([]abl.Ball, error) {
						return nil, fmt.Errorf("get approved ball list error")
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
			want: expectations{
				repoAddCalls:        0,
				repoListCalls:       0,
				repoRemoveCalls:     0,
				notifierNotifyCalls: 0,
				usbcClientGetCalls:  1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := abl.NewService(tt.fields.repo, tt.fields.notifier, tt.fields.usbcClient)
			if err := s.RefreshBalls(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("service.RefreshBalls() error = %v, wantErr %v", err, tt.wantErr)
			}

			mockRepo, _ := tt.fields.repo.(*mocks.RepositoryMock)
			mockNotifier, _ := tt.fields.notifier.(*mocks.NotifierMock)
			mockUsbcClient, _ := tt.fields.usbcClient.(*mocks.USBCClientMock)

			got := expectations{
				repoAddCalls:        len(mockRepo.AddBallsCalls()),
				repoListCalls:       len(mockRepo.ListBallsCalls()),
				repoRemoveCalls:     len(mockRepo.RemoveBallsCalls()),
				notifierNotifyCalls: len(mockNotifier.NotifyCalls()),
				usbcClientGetCalls:  len(mockUsbcClient.GetApprovedBallListCalls()),
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
