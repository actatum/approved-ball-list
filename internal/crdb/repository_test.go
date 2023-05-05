package crdb

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/actatum/approved-ball-list/internal/abl"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

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

func TestRepository_AddBalls(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	db, cleanup := StartTestDB(t, false)
	t.Cleanup(cleanup)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	t.Run("add 0 balls", func(t *testing.T) {
		cleanBallsTable(t, db)

		ctx := context.Background()
		err := repo.AddBalls(ctx, nil)
		assert.NoError(t, err, "Repository.AddBalls()")

		res, err := repo.ListBalls(ctx, abl.BallFilter{
			PageSize: 1,
		})
		require.NoError(t, err, "Repository.ListBalls()")

		assert.Equal(t, 0, len(res.Balls))
	})

	t.Run("add 1 ball", func(t *testing.T) {
		cleanBallsTable(t, db)

		ctx := context.Background()
		want := []abl.Ball{
			{
				Brand:      "Storm",
				Name:       "Fate",
				ApprovedAt: time.Now().UTC().Truncate(24 * time.Hour),
				ImageURL:   "some-url",
			},
		}
		err := repo.AddBalls(ctx, want)
		assert.NoError(t, err, "Repository.AddBalls()")

		res, err := repo.ListBalls(ctx, abl.BallFilter{
			PageSize: 1,
		})
		require.NoError(t, err, "Repository.ListBalls()")

		assert.Equal(t, 1, len(res.Balls))

		assert.Equal(t, want[0].Brand, res.Balls[0].Brand)
		assert.Equal(t, want[0].Name, res.Balls[0].Name)
		assert.Equal(t, want[0].ApprovedAt.UTC(), res.Balls[0].ApprovedAt.UTC())
		assert.Equal(t, want[0].ImageURL, res.Balls[0].ImageURL)
	})
}

func TestRepository_RemoveBalls(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	db, cleanup := StartTestDB(t, false)
	t.Cleanup(cleanup)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	t.Run("delete 0 balls", func(t *testing.T) {
		cleanBallsTable(t, db)

		ctx := context.Background()
		seedData := []abl.Ball{}
		seededData := insertSeedData(t, db, seedData...)

		if err := repo.RemoveBalls(ctx, seededData); err != nil {
			t.Fatalf("Repository.RemoveBalls() error = %v", err)
		}

		res, err := repo.ListBalls(ctx, abl.BallFilter{
			PageSize: 1,
		})
		require.NoError(t, err, "Repository.ListBalls()")

		assert.Equal(t, 0, len(res.Balls))
	})

	t.Run("delete 1 balls", func(t *testing.T) {
		cleanBallsTable(t, db)

		ctx := context.Background()
		seedData := []abl.Ball{
			{
				Brand:      "",
				Name:       "",
				ApprovedAt: time.Time{},
				ImageURL:   "",
			},
		}

		seededData := insertSeedData(t, db, seedData...)

		if err := repo.RemoveBalls(ctx, seededData); err != nil {
			t.Fatalf("Repository.RemoveBalls() error = %v", err)
		}

		res, err := repo.ListBalls(ctx, abl.BallFilter{
			PageSize: 1,
		})
		require.NoError(t, err, "Repository.ListBalls()")

		assert.Equal(t, 0, len(res.Balls))
	})
}

func TestRepository_ListBalls(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	db, cleanup := StartTestDB(t, false)
	t.Cleanup(cleanup)

	repo, err := NewRepository(db)
	require.NoError(t, err)

	t.Run("list 0 balls", func(t *testing.T) {
		cleanBallsTable(t, db)

		ctx := context.Background()
		seedData := []abl.Ball{}

		insertSeedData(t, db, seedData...)

		res, err := repo.ListBalls(ctx, abl.BallFilter{
			PageSize: 1,
		})
		require.NoError(t, err, "Repository.ListBalls()")

		assert.LessOrEqual(t, len(res.Balls), 1)

		assert.Equal(t, abl.ListBallsResult{Balls: []abl.Ball{}}, res, "Repository.ListBalls()")
	})

	t.Run("list a few balls", func(t *testing.T) {
		cleanBallsTable(t, db)

		ctx := context.Background()
		seedData := []abl.Ball{
			{
				ID:         "1",
				Brand:      "Storm",
				Name:       "Fate",
				ApprovedAt: time.Now().Truncate(1 * time.Minute),
				ImageURL:   "image",
			},
			{
				ID:         "2",
				Brand:      "Ebonite",
				Name:       "Gamebreaker 2",
				ApprovedAt: time.Now().Add(5 * time.Minute).Truncate(1 * time.Minute),
				ImageURL:   "image",
			},
			{
				ID:         "3",
				Brand:      "Motiv",
				Name:       "Sky Raptor",
				ApprovedAt: time.Now().Add(10 * time.Minute).Truncate(1 * time.Minute),
				ImageURL:   "image",
			},
		}

		insertSeedData(t, db, seedData...)

		slices.SortFunc(seedData, func(a abl.Ball, b abl.Ball) bool {
			return b.ApprovedAt.Before(a.ApprovedAt)
		})

		res, err := repo.ListBalls(ctx, abl.BallFilter{
			PageSize: 5,
		})
		require.NoError(t, err, "Repository.ListBalls()")

		assert.LessOrEqual(t, len(res.Balls), 5)

		assert.Equal(t, abl.ListBallsResult{
			Balls:         seedData,
			NextPageToken: "",
			Count:         3,
		}, res, "Repository.ListBalls()")
	})

	t.Run("list with brand filter", func(t *testing.T) {
		cleanBallsTable(t, db)

		ctx := context.Background()
		seedData := []abl.Ball{
			{
				ID:         "1",
				Brand:      "Storm",
				Name:       "Fate",
				ApprovedAt: time.Now().Truncate(1 * time.Minute),
				ImageURL:   "image",
			},
			{
				ID:         "2",
				Brand:      "Ebonite",
				Name:       "Gamebreaker 2",
				ApprovedAt: time.Now().Truncate(1 * time.Minute),
				ImageURL:   "image",
			},
			{
				ID:         "3",
				Brand:      "Motiv",
				Name:       "Sky Raptor",
				ApprovedAt: time.Now().Truncate(1 * time.Minute),
				ImageURL:   "image",
			},
		}

		insertSeedData(t, db, seedData...)

		brand := "Storm"
		res, err := repo.ListBalls(ctx, abl.BallFilter{
			PageSize: 5,
			Brand:    &brand,
		})
		require.NoError(t, err, "Repository.ListBalls()")

		assert.LessOrEqual(t, len(res.Balls), 5)
		assert.Equal(t, abl.ListBallsResult{
			Balls:         []abl.Ball{seedData[0]},
			NextPageToken: "",
			Count:         1,
		}, res, "Repository.ListBalls()")
	})

	t.Run("list with page size return next page token", func(t *testing.T) {
		cleanBallsTable(t, db)

		ctx := context.Background()
		seedData := []abl.Ball{
			{
				ID:         "1",
				Brand:      "Storm",
				Name:       "Fate",
				ApprovedAt: time.Now().Truncate(1 * time.Minute),
				ImageURL:   "image",
			},
			{
				ID:         "2",
				Brand:      "Ebonite",
				Name:       "Gamebreaker 2",
				ApprovedAt: time.Now().Add(5 * time.Minute).Truncate(1 * time.Minute),
				ImageURL:   "image",
			},
			{
				ID:         "3",
				Brand:      "Motiv",
				Name:       "Sky Raptor",
				ApprovedAt: time.Now().Add(10 * time.Minute).Truncate(1 * time.Minute),
				ImageURL:   "image",
			},
		}

		insertSeedData(t, db, seedData...)

		slices.SortFunc(seedData, func(a abl.Ball, b abl.Ball) bool {
			return b.ApprovedAt.Before(a.ApprovedAt)
		})

		res, err := repo.ListBalls(ctx, abl.BallFilter{
			PageSize: 1,
		})
		require.NoError(t, err, "Repository.ListBalls()")

		assert.LessOrEqual(t, len(res.Balls), 5)
		assert.Equal(t, abl.ListBallsResult{
			Balls:         []abl.Ball{seedData[0]},
			NextPageToken: "MQ==",
			Count:         3,
		}, res, "Repository.ListBalls()")
	})

	t.Run("list with page size and page token filter", func(t *testing.T) {
		cleanBallsTable(t, db)

		ctx := context.Background()
		seedData := []abl.Ball{
			{
				ID:         "1",
				Brand:      "Storm",
				Name:       "Fate",
				ApprovedAt: time.Now().Truncate(1 * time.Minute),
				ImageURL:   "image",
			},
			{
				ID:         "2",
				Brand:      "Ebonite",
				Name:       "Gamebreaker 2",
				ApprovedAt: time.Now().Add(5 * time.Minute).Truncate(1 * time.Minute),
				ImageURL:   "image",
			},
			{
				ID:         "3",
				Brand:      "Motiv",
				Name:       "Sky Raptor",
				ApprovedAt: time.Now().Add(10 * time.Minute).Truncate(1 * time.Minute),
				ImageURL:   "image",
			},
		}

		insertSeedData(t, db, seedData...)

		slices.SortFunc(seedData, func(a abl.Ball, b abl.Ball) bool {
			return b.ApprovedAt.Before(a.ApprovedAt)
		})

		res, err := repo.ListBalls(ctx, abl.BallFilter{
			PageSize:  2,
			PageToken: "MQ==",
		})
		require.NoError(t, err, "Repository.ListBalls()")

		assert.LessOrEqual(t, len(res.Balls), 5)
		assert.Equal(t, abl.ListBallsResult{
			Balls:         seedData[1:],
			NextPageToken: "",
			Count:         3,
		}, res, "Repository.ListBalls()")
	})
}

func cleanBallsTable(tb testing.TB, db *sqlx.DB) {
	tb.Helper()

	_, err := db.Exec("DELETE FROM balls")
	if err != nil {
		tb.Fatal(err)
	}

	_, err = db.Exec("ALTER SEQUENCE ball_ids RESTART")
	if err != nil {
		tb.Fatal(err)
	}
}

func insertSeedData(tb testing.TB, db *sqlx.DB, seedData ...abl.Ball) []abl.Ball {
	tb.Helper()

	if len(seedData) == 0 {
		return nil
	}

	batch := make([]ball, 0, len(seedData))
	for _, b := range seedData {
		batch = append(batch, ball{
			Brand:      b.Brand,
			Name:       b.Name,
			ApprovedAt: b.ApprovedAt,
			ImageURL:   b.ImageURL,
		})
	}

	var results []ball
	err := crdbsqlx.ExecuteTx(context.Background(), db, &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		_, err := tx.NamedExec(insertBallsQuery, batch)
		if err != nil {
			return fmt.Errorf("tx.NamedExec: %w", err)
		}

		err = tx.Select(&results, `SELECT * FROM balls`)
		if err != nil {
			return fmt.Errorf("tx.Select: %w", err)
		}

		return nil
	})
	if err != nil {
		tb.Fatal(err)
	}

	inserted := make([]abl.Ball, 0, len(results))
	for _, b := range results {
		inserted = append(inserted, abl.Ball{
			ID:         strconv.FormatInt(b.ID, 10),
			Brand:      b.Brand,
			Name:       b.Name,
			ApprovedAt: b.ApprovedAt,
			ImageURL:   b.ImageURL,
		})
	}

	return inserted
}
