package sqlite_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/actatum/approved-ball-list/internal/abl"
	"github.com/actatum/approved-ball-list/internal/mocks"
	"github.com/actatum/approved-ball-list/internal/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestNewRepository(t *testing.T) {
	type args struct {
		url string
		bm  sqlite.BackupManager
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success (file)",
			args: args{
				url: "sqlite.db",
				bm: &mocks.BackupManagerMock{
					BackupFunc: func(ctx context.Context, file string) error {
						return nil
					},
					RestoreFunc: func(ctx context.Context, file string) error {
						return nil
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success (memory)",
			args: args{
				url: ":memory:",
				bm: &mocks.BackupManagerMock{
					BackupFunc: func(ctx context.Context, file string) error {
						return nil
					},
					RestoreFunc: func(ctx context.Context, file string) error {
						return nil
					},
				},
			},
			wantErr: false,
		},
		{
			name: "restore error",
			args: args{
				url: "sqlite.db",
				bm: &mocks.BackupManagerMock{
					BackupFunc: func(ctx context.Context, file string) error {
						return nil
					},
					RestoreFunc: func(ctx context.Context, file string) error {
						return fmt.Errorf("restore error")
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sqlite.NewRepository(tt.args.url, tt.args.bm)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Cleanup(func() {
				if tt.args.url != ":memory" {
					_ = os.Remove(tt.args.url)
				}

				if got != nil {
					_ = got.Close()
				}
			})
		})
	}
}

func TestRepository_AddBalls(t *testing.T) {
	type fields struct {
		url string
		bm  sqlite.BackupManager
	}
	type args struct {
		ctx   context.Context
		balls []abl.Ball
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "add 0 balls",
			fields: fields{
				bm: &mocks.BackupManagerMock{
					BackupFunc: func(ctx context.Context, file string) error {
						return nil
					},
					RestoreFunc: func(ctx context.Context, file string) error {
						return nil
					},
				},
				url: ":memory:",
			},
			args: args{
				ctx:   context.Background(),
				balls: nil,
			},
			wantErr: false,
		},
		{
			name: "add 1 ball",
			fields: fields{
				bm: &mocks.BackupManagerMock{
					BackupFunc: func(ctx context.Context, file string) error {
						return nil
					},
					RestoreFunc: func(ctx context.Context, file string) error {
						return nil
					},
				},
				url: ":memory:",
			},
			args: args{
				ctx: context.Background(),
				balls: []abl.Ball{
					{
						ID:         "1",
						Brand:      "Storm",
						Name:       "Fate",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "some-url",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := sqlite.NewRepository(tt.fields.url, tt.fields.bm)
			require.NoError(t, err, "sqlite.NewRepository()")

			t.Cleanup(func() {
				err := r.Close()
				require.NoError(t, err)
			})

			if err := r.AddBalls(tt.args.ctx, tt.args.balls); (err != nil) != tt.wantErr {
				t.Errorf("Repository.AddBalls() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			res, err := r.ListBalls(tt.args.ctx, abl.BallFilter{
				PageSize: len(tt.args.balls),
			})
			require.NoError(t, err, "Repository.ListBalls()")

			assert.Equal(t, len(tt.args.balls), len(res.Balls))

			for _, b := range tt.args.balls {
				assert.Contains(t, res.Balls, b, "expected result to contain: %v", b)
			}
		})
	}
}

func TestRepository_RemoveBalls(t *testing.T) {
	type fields struct {
		url string
		bm  sqlite.BackupManager
	}
	type args struct {
		ctx      context.Context
		balls    []abl.Ball
		seedData []abl.Ball
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "delete 0 balls",
			fields: fields{
				bm: &mocks.BackupManagerMock{
					BackupFunc: func(ctx context.Context, file string) error {
						return nil
					},
					RestoreFunc: func(ctx context.Context, file string) error {
						return nil
					},
				},
				url: ":memory:",
			},
			args: args{
				ctx:      context.Background(),
				balls:    nil,
				seedData: nil,
			},
			wantErr: false,
		},
		{
			name: "delete 2 balls",
			fields: fields{
				bm: &mocks.BackupManagerMock{
					BackupFunc: func(ctx context.Context, file string) error {
						return nil
					},
					RestoreFunc: func(ctx context.Context, file string) error {
						return nil
					},
				},
				url: ":memory:",
			},
			args: args{
				ctx: context.Background(),
				balls: []abl.Ball{
					{
						Brand:      "Storm",
						Name:       "Fate",
						ApprovedAt: time.Now(),
						ImageURL:   "image-url",
					},
					{
						Brand:      "Storm",
						Name:       "Super Nova",
						ApprovedAt: time.Now().Add(-1 * 24 * time.Hour),
						ImageURL:   "image-url",
					},
				},
				seedData: []abl.Ball{
					{
						Brand:      "Storm",
						Name:       "Fate",
						ApprovedAt: time.Now(),
						ImageURL:   "image-url",
					},
					{
						Brand:      "Storm",
						Name:       "Super Nova",
						ApprovedAt: time.Now().Add(-1 * 24 * time.Hour),
						ImageURL:   "image-url",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := sqlite.NewRepository(tt.fields.url, tt.fields.bm)
			require.NoError(t, err, "sqlite.NewRepository()")

			t.Cleanup(func() {
				err := r.Close()
				require.NoError(t, err)
			})

			err = r.AddBalls(tt.args.ctx, tt.args.seedData)
			require.NoError(t, err, "Repository.AddBalls()")

			if err := r.RemoveBalls(tt.args.ctx, tt.args.balls); (err != nil) != tt.wantErr {
				t.Fatalf("Repository.RemoveBalls() error = %v, wantErr %v", err, tt.wantErr)
			}

			res, err := r.ListBalls(tt.args.ctx, abl.BallFilter{
				PageSize: len(tt.args.balls),
			})
			require.NoError(t, err, "Repository.ListBalls()")

			assert.Equal(t, len(tt.args.seedData)-len(tt.args.balls), len(res.Balls))

			for _, b := range tt.args.balls {
				assert.NotContains(t, res.Balls, b, "expected result not to contain: %v", b)
			}
		})
	}
}

func TestRepository_ListBalls(t *testing.T) {
	type fields struct {
		url string
		bm  sqlite.BackupManager
	}
	type args struct {
		ctx      context.Context
		filter   abl.BallFilter
		seedData []abl.Ball
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    abl.ListBallsResult
		wantErr bool
	}{
		{
			name: "list 0 balls",
			fields: fields{
				bm: &mocks.BackupManagerMock{
					BackupFunc: func(ctx context.Context, file string) error {
						return nil
					},
					RestoreFunc: func(ctx context.Context, file string) error {
						return nil
					},
				},
				url: ":memory:",
			},
			args: args{
				ctx:      context.Background(),
				filter:   abl.BallFilter{},
				seedData: nil,
			},
			want: abl.ListBallsResult{
				Balls:         []abl.Ball{},
				NextPageToken: "",
				Count:         0,
			},
			wantErr: false,
		},
		{
			name: "list a few balls",
			fields: fields{
				bm: &mocks.BackupManagerMock{
					BackupFunc: func(ctx context.Context, file string) error {
						return nil
					},
					RestoreFunc: func(ctx context.Context, file string) error {
						return nil
					},
				},
				url: ":memory:",
			},
			args: args{
				ctx: context.Background(),
				filter: abl.BallFilter{
					PageSize: 3,
				},
				seedData: []abl.Ball{
					{
						ID:         "1",
						Brand:      "Storm",
						Name:       "Fate",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
					{
						ID:         "2",
						Brand:      "Ebonite",
						Name:       "Gamebreaker 2",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
					{
						ID:         "3",
						Brand:      "Motiv",
						Name:       "Sky Raptor",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
				},
			},
			want: abl.ListBallsResult{
				Balls: []abl.Ball{
					{
						ID:         "1",
						Brand:      "Storm",
						Name:       "Fate",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
					{
						ID:         "2",
						Brand:      "Ebonite",
						Name:       "Gamebreaker 2",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
					{
						ID:         "3",
						Brand:      "Motiv",
						Name:       "Sky Raptor",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
				},
				NextPageToken: "",
				Count:         3,
			},
			wantErr: false,
		},
		{
			name: "list with brand filter",
			fields: fields{
				bm: &mocks.BackupManagerMock{
					BackupFunc: func(ctx context.Context, file string) error {
						return nil
					},
					RestoreFunc: func(ctx context.Context, file string) error {
						return nil
					},
				},
				url: ":memory:",
			},
			args: args{
				ctx: context.Background(),
				filter: abl.BallFilter{
					Brand:    stringPtr("Storm"),
					PageSize: 3,
				},
				seedData: []abl.Ball{
					{
						ID:         "1",
						Brand:      "Storm",
						Name:       "Fate",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
					{
						ID:         "2",
						Brand:      "Ebonite",
						Name:       "Gamebreaker 2",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
					{
						ID:         "3",
						Brand:      "Motiv",
						Name:       "Sky Raptor",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
				},
			},
			want: abl.ListBallsResult{
				Balls: []abl.Ball{
					{
						ID:         "1",
						Brand:      "Storm",
						Name:       "Fate",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
				},
				NextPageToken: "",
				Count:         1,
			},
			wantErr: false,
		},
		{
			name: "list with page size filter return next page token",
			fields: fields{
				bm: &mocks.BackupManagerMock{
					BackupFunc: func(ctx context.Context, file string) error {
						return nil
					},
					RestoreFunc: func(ctx context.Context, file string) error {
						return nil
					},
				},
				url: ":memory:",
			},
			args: args{
				ctx: context.Background(),
				filter: abl.BallFilter{
					PageSize: 1,
				},
				seedData: []abl.Ball{
					{
						ID:         "1",
						Brand:      "Storm",
						Name:       "Fate",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
					{
						ID:         "2",
						Brand:      "Ebonite",
						Name:       "Gamebreaker 2",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
					{
						ID:         "3",
						Brand:      "Motiv",
						Name:       "Sky Raptor",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
				},
			},
			want: abl.ListBallsResult{
				Balls: []abl.Ball{
					{
						ID:         "1",
						Brand:      "Storm",
						Name:       "Fate",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
				},
				NextPageToken: "MQ==",
				Count:         3,
			},
			wantErr: false,
		},
		{
			name: "list with page size and page token filter",
			fields: fields{
				bm: &mocks.BackupManagerMock{
					BackupFunc: func(ctx context.Context, file string) error {
						return nil
					},
					RestoreFunc: func(ctx context.Context, file string) error {
						return nil
					},
				},
				url: ":memory:",
			},
			args: args{
				ctx: context.Background(),
				filter: abl.BallFilter{
					PageToken: "MQ==",
					PageSize:  2,
				},
				seedData: []abl.Ball{
					{
						ID:         "1",
						Brand:      "Storm",
						Name:       "Fate",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
					{
						ID:         "2",
						Brand:      "Ebonite",
						Name:       "Gamebreaker 2",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
					{
						ID:         "3",
						Brand:      "Motiv",
						Name:       "Sky Raptor",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
				},
			},
			want: abl.ListBallsResult{
				Balls: []abl.Ball{
					{
						ID:         "2",
						Brand:      "Ebonite",
						Name:       "Gamebreaker 2",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
					{
						ID:         "3",
						Brand:      "Motiv",
						Name:       "Sky Raptor",
						ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
						ImageURL:   "image",
					},
				},
				NextPageToken: "",
				Count:         3,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := sqlite.NewRepository(tt.fields.url, tt.fields.bm)
			require.NoError(t, err, "sqlite.NewRepository()")

			t.Cleanup(func() {
				err := r.Close()
				require.NoError(t, err)
			})

			err = r.AddBalls(tt.args.ctx, tt.args.seedData)
			require.NoError(t, err, "Repository.AddBalls()")

			res, err := r.ListBalls(tt.args.ctx, tt.args.filter)
			require.NoError(t, err, "Repository.ListBalls()")

			assert.LessOrEqual(t, len(res.Balls), tt.args.filter.PageSize)

			assert.Equal(t, tt.want, res, "Repository.ListBalls()")
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
