package cockroachdb

import (
	"context"
	"testing"
	"time"

	"github.com/actatum/approved-ball-list/abl"
)

func TestBallService_AddBalls(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()

		s := NewBallService(db)

		b := abl.Ball{
			Brand:      "Storm",
			Name:       "Phaze II",
			ApprovedAt: time.Now(),
			ImageURL:   "abc",
		}

		if err := s.AddBalls(ctx, b); err != nil {
			t.Fatal(err)
		}

		row := db.QueryRowContext(ctx, `SELECT * FROM balls WHERE id = 1`)
		if row.Err() != nil {
			t.Fatal(row.Err())
		}

		var fromDB abl.Ball
		if err := row.Scan(
			&fromDB.ID,
			&fromDB.Brand,
			&fromDB.Name,
			&fromDB.ImageURL,
			&fromDB.ApprovedAt,
		); err != nil {
			t.Fatal(err)
		}

		if fromDB.ID != 1 {
			t.Fatalf("ID=%v, want %v", fromDB.ID, 1)
		} else if fromDB.Brand != b.Brand {
			t.Fatalf("Brand=%v, want %v", fromDB.Brand, b.Brand)
		} else if fromDB.Name != b.Name {
			t.Fatalf("Name=%v, want %v", fromDB.Name, b.Name)
		} else if fromDB.ImageURL != b.ImageURL {
			t.Fatalf("ImageURL=%v, want %v", fromDB.ImageURL, b.ImageURL)
		}
	})
}

func TestBallService_ListBalls(t *testing.T) {
	t.Parallel()

	t.Run("OK no filters", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()

		s := NewBallService(db)

		b := abl.Ball{
			Brand:      "Storm",
			Name:       "Phaze II",
			ApprovedAt: time.Now(),
			ImageURL:   "abc",
		}

		if err := s.AddBalls(ctx, b); err != nil {
			t.Fatal(err)
		}

		list, cnt, err := s.ListBalls(ctx, abl.BallFilter{})
		if err != nil {
			t.Fatal(err)
		}

		if cnt != 1 {
			t.Fatalf("cnt=%v, want %v", cnt, 1)
		}

		if list[0].ID != 1 {
			t.Fatalf("ID=%v, want %v", list[0].ID, 1)
		} else if list[0].Brand != b.Brand {
			t.Fatalf("Brand=%v, want %v", list[0].Brand, b.Brand)
		} else if list[0].Name != b.Name {
			t.Fatalf("Name=%v, want %v", list[0].Name, b.Name)
		} else if list[0].ImageURL != b.ImageURL {
			t.Fatalf("ImageURL=%v, want %v", list[0].ImageURL, b.ImageURL)
		}
	})

	t.Run("OK w/filters", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()

		s := NewBallService(db)

		b1 := abl.Ball{
			Brand:      "Storm",
			Name:       "Phaze II",
			ApprovedAt: time.Now(),
			ImageURL:   "abc",
		}

		b2 := abl.Ball{
			Brand:      "Hammer",
			Name:       "Black Widow Ghost",
			ApprovedAt: time.Now(),
			ImageURL:   "def",
		}

		if err := s.AddBalls(ctx, b1, b2); err != nil {
			t.Fatal(err)
		}

		brand := "Storm"
		list, cnt, err := s.ListBalls(ctx, abl.BallFilter{Brand: &brand})
		if err != nil {
			t.Fatal(err)
		}

		if cnt != 1 {
			t.Fatalf("cnt=%v, want %v", cnt, 1)
		}

		if list[0].ID != 1 {
			t.Fatalf("ID=%v, want %v", list[0].ID, 1)
		} else if list[0].Brand != b1.Brand {
			t.Fatalf("Brand=%v, want %v", list[0].Brand, b1.Brand)
		} else if list[0].Name != b1.Name {
			t.Fatalf("Name=%v, want %v", list[0].Name, b1.Name)
		} else if list[0].ImageURL != b1.ImageURL {
			t.Fatalf("ImageURL=%v, want %v", list[0].ImageURL, b1.ImageURL)
		}
	})
}

func TestBallService_RemoveBalls(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		ctx := context.Background()

		s := NewBallService(db)

		b1 := abl.Ball{
			Brand:      "Storm",
			Name:       "Phaze II",
			ApprovedAt: time.Now(),
			ImageURL:   "abc",
		}

		b2 := abl.Ball{
			Brand:      "Hammer",
			Name:       "Black Widow Ghost",
			ApprovedAt: time.Now(),
			ImageURL:   "def",
		}

		if err := s.AddBalls(ctx, b1, b2); err != nil {
			t.Fatal(err)
		}

		_, cnt, err := s.ListBalls(ctx, abl.BallFilter{})
		if err != nil {
			t.Fatal(err)
		}

		if cnt != 2 {
			t.Fatalf("cnt=%v, want %v", cnt, 2)
		}

		if err = s.RemoveBalls(ctx, b1, b2); err != nil {
			t.Fatal(err)
		}

		_, cnt, err = s.ListBalls(ctx, abl.BallFilter{})
		if err != nil {
			t.Fatal(err)
		}

		if cnt != 0 {
			t.Fatalf("cnt=%v, want %v", cnt, 0)
		}
	})
}
