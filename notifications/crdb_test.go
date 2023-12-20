package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/actatum/approved-ball-list/balls"
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

func TestCRDBRepository_FindAllTargets(t *testing.T) {
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
			got, err := r.FindAllTargets(tt.args.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("FindAllTargets(%v)", tt.args.ctx)) {
				return
			}
			assert.Equalf(t, tt.want, got, "FindAllTargets(%v)", tt.args.ctx)
		})
	}
}

func TestCRDBRepository_Store(t *testing.T) {
	t.Parallel()

	db, close := crdbtest.StartTestContainer(t)
	t.Cleanup(close)

	type args struct {
		ctx           context.Context
		notifications []Notification
	}
	tests := []struct {
		name              string
		args              args
		seedTargets       []Target
		seedNotifications []Notification
		wantErr           assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				notifications: []Notification{
					{
						ID:    "a72260f8-21e6-48c9-9fb3-2f19f9d649aa",
						State: StatePending,
						Content: []balls.Ball{
							{
								ID:           "2",
								Brand:        balls.Storm,
								Name:         "Phaze II",
								ApprovalDate: time.Now().Truncate(1 * time.Minute),
								ImageURL:     "https://image.url",
							},
						},
						Target: Target{
							ID: "1b7f4899-ae10-46ab-b73e-47310cf62e16",
						},
						SentAt: time.Time{},
					},
				},
			},
			seedTargets: []Target{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanNotificationsTable(t, db)
			cleanTargetsTable(t, db)
			seedTargets(t, db, tt.seedTargets...)
			seedNotifications(t, db, tt.seedNotifications...)
			r := &CRDBRepository{
				db: db,
			}
			err := r.Store(tt.args.ctx, tt.args.notifications)
			if !tt.wantErr(t, err, fmt.Sprintf("Store(%v, %v)", tt.args.ctx, tt.args.notifications)) {
				return
			}
		})
	}
}

func cleanNotificationsTable(tb testing.TB, db *sql.DB) {
	tb.Helper()

	_, err := db.Exec(`DELETE FROM notifications`)
	if err != nil {
		tb.Fatal(err)
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

func seedNotifications(tb testing.TB, db *sql.DB, notifications ...Notification) {
	tb.Helper()

	for _, notif := range notifications {
		ballID, err := strconv.ParseInt(notif.Content[0].ID, 10, 64)
		if err != nil {
			tb.Fatal(err)
		}

		_, err = db.Exec(
			`INSERT INTO notifications (id, state, ball_id, target_id, sent_at) VALUES ($1, $2, $3, $4, $5)`,
			notif.ID, notif.State, ballID, notif.Target.ID, notif.SentAt,
		)
		if err != nil {
			tb.Fatal(err)
		}
	}
}
