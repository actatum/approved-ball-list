package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/actatum/approved-ball-list/crdbtest"
	"github.com/stretchr/testify/assert"
)

func TestCRDBRepository_StoreTarget(t *testing.T) {
	t.Parallel()

	db, close := crdbtest.StartTestContainer(t)
	t.Cleanup(close)

	type args struct {
		ctx    context.Context
		target Target
	}
	tests := []struct {
		name    string
		args    args
		seed    []Target
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				target: Target{
					ID:          "5ac8f91d-8546-492b-bae6-84948a4cd603",
					Type:        TargetTypeDiscord,
					Destination: "some_channel_id",
					CreatedAt:   time.Now().Truncate(1 * time.Minute),
					UpdateAt:    time.Now().Truncate(1 * time.Minute),
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "duplicate target id",
			args: args{
				ctx: context.Background(),
				target: Target{
					ID:          "5ac8f91d-8546-492b-bae6-84948a4cd603",
					Type:        TargetTypeDiscord,
					Destination: "some_channel_id",
					CreatedAt:   time.Now().Truncate(1 * time.Minute),
					UpdateAt:    time.Now().Truncate(1 * time.Minute),
				},
			},
			seed: []Target{
				{
					ID:          "5ac8f91d-8546-492b-bae6-84948a4cd603",
					Type:        TargetTypeDiscord,
					Destination: "some_channel_id",
					CreatedAt:   time.Now().Truncate(1 * time.Minute),
					UpdateAt:    time.Now().Truncate(1 * time.Minute),
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "duplicate target destination",
			args: args{
				ctx: context.Background(),
				target: Target{
					ID:          "57ba9a19-1208-46d1-910e-b58a33eb36a1",
					Type:        TargetTypeDiscord,
					Destination: "some_channel_id",
					CreatedAt:   time.Now().Truncate(1 * time.Minute),
					UpdateAt:    time.Now().Truncate(1 * time.Minute),
				},
			},
			seed: []Target{
				{
					ID:          "5ac8f91d-8546-492b-bae6-84948a4cd603",
					Type:        TargetTypeDiscord,
					Destination: "some_channel_id",
					CreatedAt:   time.Now().Truncate(1 * time.Minute),
					UpdateAt:    time.Now().Truncate(1 * time.Minute),
				},
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorAs(tt, err, &DuplicateTargetError{})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanTargetsTable(t, db)
			seedTargets(t, db, tt.seed...)
			r := &CRDBRepository{
				db: db,
			}
			err := r.StoreTarget(tt.args.ctx, tt.args.target)
			if !tt.wantErr(t, err, fmt.Sprintf("StoreTarget(%v, %v)", tt.args.ctx, tt.args.target)) {
				return
			}
		})
	}
}

func TestCRDBRepository_FindAll(t *testing.T) {
	t.Parallel()

	db, close := crdbtest.StartTestContainer(t)
	t.Cleanup(close)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		seed    []Target
		want    []Target
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
			},
			seed: []Target{
				{
					ID:          "5ac8f91d-8546-492b-bae6-84948a4cd603",
					Type:        TargetTypeDiscord,
					Destination: "some_channel_id",
					CreatedAt:   time.Now().Truncate(1 * time.Minute),
					UpdateAt:    time.Now().Truncate(1 * time.Minute),
				},
			},
			want: []Target{
				{
					ID:          "5ac8f91d-8546-492b-bae6-84948a4cd603",
					Type:        TargetTypeDiscord,
					Destination: "some_channel_id",
					CreatedAt:   time.Now().Truncate(1 * time.Minute),
					UpdateAt:    time.Now().Truncate(1 * time.Minute),
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "found none",
			args: args{
				ctx: context.Background(),
			},
			seed:    []Target{},
			want:    []Target{},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanTargetsTable(t, db)
			seedTargets(t, db, tt.seed...)
			r := &CRDBRepository{
				db: db,
			}
			got, err := r.FindAll(tt.args.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("FindAll(%v)", tt.args.ctx)) {
				return
			}
			assert.Equalf(t, tt.want, got, "FindAll(%v)", tt.args.ctx)
		})
	}
}

func cleanTargetsTable(tb testing.TB, db *sql.DB) {
	tb.Helper()

	_, err := db.Exec(`DELETE FROM notification_targets`)
	if err != nil {
		tb.Fatal(err)
	}
}

func seedTargets(tb testing.TB, db *sql.DB, targets ...Target) {
	tb.Helper()

	for _, target := range targets {
		_, err := db.Exec(
			`INSERT INTO notification_targets (id, created_at, updated_at, type, destination)
		VALUES ($1, $2, $3, $4, $5)`,
			target.ID, target.CreatedAt, target.UpdateAt, target.Type, target.Destination,
		)
		if err != nil {
			tb.Fatal(err)
		}
	}
}
