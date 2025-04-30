package crdb

import (
	"context"
	"database/sql"
	"net/url"
	"testing"
	"time"

	"github.com/actatum/approved-ball-list/internal/balls"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	// imported for side effects
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestNewRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	db, cleanup := StartTestDB(t, false)
	t.Cleanup(cleanup)

	_, err := NewRepository(db)
	assert.NoError(t, err)
}

func TestAddBalls(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, cleanup := StartTestDB(t, false)
	t.Cleanup(cleanup)

	repo, err := NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add 1 ball", func(t *testing.T) {
		cleanBallsTable(t, db)

		want := []balls.Ball{
			{
				ID:           1,
				Brand:        "Storm",
				Name:         "Phaze II",
				ApprovalDate: time.Now(),
				ImageURL:     &url.URL{},
			},
		}

		got, err := repo.Add(context.Background(), want...)
		if err != nil {
			t.Errorf("repo.Add() error = %v", err)
		}

		if diff := cmp.Diff(want, got, cmpopts.EquateApproxTime(time.Second)); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("ball already exists", func(t *testing.T) {
		cleanBallsTable(t, db)
		seedBalls(t, db, balls.Ball{
			ID:           1,
			Brand:        "Storm",
			Name:         "Phaze II",
			ApprovalDate: time.Now().Truncate(24 * time.Hour),
			ImageURL:     &url.URL{},
		})

		input := []balls.Ball{
			{
				ID:           1,
				Brand:        "Storm",
				Name:         "Phaze II",
				ApprovalDate: time.Now().Truncate(24 * time.Hour),
				ImageURL:     &url.URL{},
			},
		}

		got, err := repo.Add(context.Background(), input...)
		if err != nil {
			t.Errorf("repo.Add() error = %v", err)
		}

		if len(got) != 0 {
			t.Errorf("expected 0 balls, got %d", len(got))
		}
	})
}

func cleanBallsTable(tb testing.TB, db *sql.DB) {
	tb.Helper()

	if _, err := db.Exec("DELETE FROM balls"); err != nil {
		tb.Fatal(err)
	}
}

func seedBalls(tb testing.TB, db *sql.DB, seed ...balls.Ball) {
	tb.Helper()

	for _, b := range seed {
		if _, err := db.Exec(
			"INSERT INTO balls (brand, name, image_url, approved_at) VALUES ($1, $2, $3, $4)",
			b.Brand, b.Name, b.ImageURL, b.ApprovalDate,
		); err != nil {
			tb.Fatal(err)
		}
	}
}
