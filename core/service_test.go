package core

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockAlerter struct {
	mock.Mock
}

func (m *mockAlerter) SendMessage(_ context.Context, _ []string, _ Ball) error {
	args := m.Called()
	return args.Error(0)
}

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) GetAllBalls(_ context.Context) ([]Ball, error) {
	args := m.Called()
	return args.Get(0).([]Ball), args.Error(1)
}

func (m *mockRepository) InsertNewBalls(_ context.Context, _ []Ball) error {
	args := m.Called()
	return args.Error(0)
}

type mockUSBC struct {
	mock.Mock
}

func (m *mockUSBC) GetApprovedBallList(_ context.Context) ([]Ball, error) {
	args := m.Called()
	return args.Get(0).([]Ball), args.Error(1)
}

func Test_service_FilterAndAddBalls(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "no filter just add",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := new(mockAlerter)
			mr := new(mockRepository)
			mu := new(mockUSBC)
			logger, err := zap.NewProduction()
			if err != nil {
				t.Fatalf("failed to init logger: %v", err)
			}
			s := NewService(&Config{
				Alerter:         ma,
				Repository:      mr,
				USBC:            mu,
				Logger:          logger,
				DiscordChannels: nil,
			})

			mr.On("GetAllBalls").Return([]Ball{}, nil)
			mu.On("GetApprovedBallList").Return([]Ball{}, nil)
			mr.On("InsertNewBalls").Return(nil)

			if err := s.FilterAndAddBalls(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("service.FilterAndAddBalls() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_service_AlertNewBall(t *testing.T) {
	type args struct {
		ctx  context.Context
		ball Ball
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "new ball alert",
			args: args{
				ctx: context.Background(),
				ball: Ball{
					Brand: "Motiv",
					Name:  "Iron Forge",
				},
			},
			wantErr: false,
		},
		{
			name: "alerter error",
			args: args{
				ctx:  context.Background(),
				ball: Ball{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := new(mockAlerter)
			mr := new(mockRepository)
			mu := new(mockUSBC)
			logger, err := zap.NewProduction()
			if err != nil {
				t.Fatalf("failed to init logger: %v", err)
			}
			s := NewService(&Config{
				Alerter:    ma,
				Repository: mr,
				USBC:       mu,
				Logger:     logger,
				DiscordChannels: map[string]DiscordChannel{
					"motivated": {
						ID:     "1",
						Name:   "motivated",
						Brands: []string{"Motiv"},
					},
				},
			})

			if !tt.wantErr {
				ma.On("SendMessage").Return(nil)
			} else {
				ma.On("SendMessage").Return(errors.New("error"))
			}

			if err := s.AlertNewBall(tt.args.ctx, tt.args.ball); (err != nil) != tt.wantErr {
				t.Errorf("service.AlertNewBall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_service_filter(t *testing.T) {
	type args struct {
		fromRepo []Ball
		fromUSBC []Ball
	}
	tests := []struct {
		name string
		args args
		want []Ball
	}{
		// TODO: Add test cases.
		{
			name: "filter balls",
			args: args{
				fromRepo: []Ball{
					{
						Brand: "Storm",
						Name:  "Spectre",
					},
				},
				fromUSBC: []Ball{
					{
						Brand: "Storm",
						Name:  "Spectre",
					},
					{
						Brand: "Storm",
						Name:  "Nova",
					},
				},
			},
			want: []Ball{
				{
					Brand: "Storm",
					Name:  "Nova",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ma := new(mockAlerter)
			mr := new(mockRepository)
			mu := new(mockUSBC)
			logger, err := zap.NewProduction()
			if err != nil {
				t.Fatalf("failed to init logger: %v", err)
			}
			s := service{
				alerter:         ma,
				repository:      mr,
				usbc:            mu,
				logger:          logger,
				discordChannels: nil,
			}

			got := s.filter(tt.args.fromRepo, tt.args.fromUSBC)
			assert.Equal(t, tt.want, got, "service.filter()")
		})
	}
}
