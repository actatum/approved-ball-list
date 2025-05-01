package balls

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/actatum/approved-ball-list/internal/crdb"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestCRDBStore_AddBalls(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		db, cleanup := crdb.StartTestDB(t, false)
		t.Cleanup(cleanup)

		input := []Ball{
			{
				Brand: Storm,
				Name:  "Phaze II",
				ImageURL: &url.URL{
					Scheme: "http",
					Host:   "some-url",
				},
				ApprovalDate: time.Now(),
			},
			{
				Brand: Motiv,
				Name:  "Venom Shock",
				ImageURL: &url.URL{
					Scheme: "http",
					Host:   "some-url",
				},
				ApprovalDate: time.Now(),
			},
		}

		ctx := context.Background()
		s := NewCRDBStore(db)

		err := s.AddBalls(ctx, input)
		if err != nil {
			t.Fatal(err)
		}

		stmt := `SELECT id, brand, name, image_url, approved_at FROM balls`
		rows, err := db.Query(ctx, stmt)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		var got []Ball
		for rows.Next() {
			var g Ball
			var imageURL string
			err = rows.Scan(&g.ID, &g.Brand, &g.Name, &imageURL, &g.ApprovalDate)
			if err != nil {
				t.Fatal(err)
			}

			g.ImageURL, err = url.Parse(imageURL)
			if err != nil {
				t.Fatal(err)
			}

			got = append(got, g)
		}

		if len(got) != 2 {
			t.Fatalf("expected two balls got %d", len(got))
		}

		diff := cmp.Diff(got[0], input[0], cmpopts.EquateApproxTime(time.Second), cmpopts.IgnoreFields(Ball{}, "ID"))
		if diff != "" {
			t.Fatalf("(-got, +want):\n%s", diff)
		}
		if got[0].ID == 0 {
			t.Fatalf("expected db to set id")
		}

		diff = cmp.Diff(got[1], input[1], cmpopts.EquateApproxTime(time.Second), cmpopts.IgnoreFields(Ball{}, "ID"))
		if diff != "" {
			t.Fatalf("(-got, +want):\n%s", diff)
		}
		if got[1].ID == 0 {
			t.Fatalf("expected db to set id")
		}
	})

	t.Run("duplicate brand, name, and approved_at", func(t *testing.T) {
		t.Parallel()

		db, cleanup := crdb.StartTestDB(t, false)
		t.Cleanup(cleanup)

		now := time.Now()
		seed := Ball{
			Brand: Hammer,
			Name:  "Black Widow Mania",
			ImageURL: &url.URL{
				Scheme: "http",
				Host:   "some-url",
			},
			ApprovalDate: now,
		}

		ctx := context.Background()

		stmt := `INSERT INTO balls (brand, name, image_url, approved_at) VALUES ($1, $2, $3, $4)`
		_, err := db.Exec(ctx, stmt, seed.Brand, seed.Name, seed.ImageURL, seed.ApprovalDate)
		if err != nil {
			t.Fatal(err)
		}

		input := []Ball{seed}

		s := NewCRDBStore(db)

		err = s.AddBalls(ctx, input)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
	})
}

func TestCRDBStore_GetAllBalls(t *testing.T) {
	t.Parallel()

	t.Run("success no filters", func(t *testing.T) {
		t.Parallel()

		db, cleanup := crdb.StartTestDB(t, false)
		t.Cleanup(cleanup)

		now := time.Now()
		seed := []Ball{
			{
				Brand: Hammer,
				Name:  "Black Widow Mania",
				ImageURL: &url.URL{
					Scheme: "http",
					Host:   "some-url",
				},
				ApprovalDate: now,
			},
			{
				Brand: Ebonite,
				Name:  "The One Reverb",
				ImageURL: &url.URL{
					Scheme: "http",
					Host:   "some-url",
				},
				ApprovalDate: now,
			},
		}

		ctx := context.Background()

		for _, b := range seed {
			stmt := `INSERT INTO balls (brand, name, image_url, approved_at) VALUES ($1, $2, $3, $4)`
			_, err := db.Exec(ctx, stmt, b.Brand, b.Name, b.ImageURL, b.ApprovalDate)
			if err != nil {
				t.Fatal(err)
			}
		}

		s := NewCRDBStore(db)

		got, err := s.GetAllBalls(ctx, BallFilter{})
		if err != nil {
			t.Fatal("expected error but got nil")
		}

		if len(got) != 2 {
			t.Fatalf("expected two balls got %d", len(got))
		}

		diff := cmp.Diff(got[0], seed[0], cmpopts.EquateApproxTime(time.Second), cmpopts.IgnoreFields(Ball{}, "ID"))
		if diff != "" {
			t.Fatalf("(-got, +want):\n%s", diff)
		}
		if got[0].ID == 0 {
			t.Fatalf("expected db to set id")
		}

		diff = cmp.Diff(got[1], seed[1], cmpopts.EquateApproxTime(time.Second), cmpopts.IgnoreFields(Ball{}, "ID"))
		if diff != "" {
			t.Fatalf("(-got, +want):\n%s", diff)
		}
		if got[1].ID == 0 {
			t.Fatalf("expected db to set id")
		}
	})

	t.Run("success brand filter", func(t *testing.T) {
		t.Parallel()

		db, cleanup := crdb.StartTestDB(t, false)
		t.Cleanup(cleanup)

		now := time.Now()
		seed := []Ball{
			{
				Brand: Hammer,
				Name:  "Black Widow Mania",
				ImageURL: &url.URL{
					Scheme: "http",
					Host:   "some-url",
				},
				ApprovalDate: now,
			},
			{
				Brand: Ebonite,
				Name:  "The One Reverb",
				ImageURL: &url.URL{
					Scheme: "http",
					Host:   "some-url",
				},
				ApprovalDate: now,
			},
		}

		ctx := context.Background()

		for _, b := range seed {
			stmt := `INSERT INTO balls (brand, name, image_url, approved_at) VALUES ($1, $2, $3, $4)`
			_, err := db.Exec(ctx, stmt, b.Brand, b.Name, b.ImageURL, b.ApprovalDate)
			if err != nil {
				t.Fatal(err)
			}
		}

		s := NewCRDBStore(db)

		got, err := s.GetAllBalls(ctx, BallFilter{
			Brand: &seed[0].Brand,
		})
		if err != nil {
			t.Fatal("expected error but got nil")
		}

		if len(got) != 1 {
			t.Fatalf("expected one balls got %d", len(got))
		}

		diff := cmp.Diff(got[0], seed[0], cmpopts.EquateApproxTime(time.Second), cmpopts.IgnoreFields(Ball{}, "ID"))
		if diff != "" {
			t.Fatalf("(-got, +want):\n%s", diff)
		}
		if got[0].ID == 0 {
			t.Fatalf("expected db to set id")
		}
	})

	t.Run("success name filter", func(t *testing.T) {
		t.Parallel()

		db, cleanup := crdb.StartTestDB(t, false)
		t.Cleanup(cleanup)

		now := time.Now()
		seed := []Ball{
			{
				Brand: Hammer,
				Name:  "Black Widow Mania",
				ImageURL: &url.URL{
					Scheme: "http",
					Host:   "some-url",
				},
				ApprovalDate: now,
			},
			{
				Brand: Ebonite,
				Name:  "The One Reverb",
				ImageURL: &url.URL{
					Scheme: "http",
					Host:   "some-url",
				},
				ApprovalDate: now,
			},
		}

		ctx := context.Background()

		for _, b := range seed {
			stmt := `INSERT INTO balls (brand, name, image_url, approved_at) VALUES ($1, $2, $3, $4)`
			_, err := db.Exec(ctx, stmt, b.Brand, b.Name, b.ImageURL, b.ApprovalDate)
			if err != nil {
				t.Fatal(err)
			}
		}

		s := NewCRDBStore(db)

		got, err := s.GetAllBalls(ctx, BallFilter{
			Name: &seed[1].Name,
		})
		if err != nil {
			t.Fatal("expected error but got nil")
		}

		if len(got) != 1 {
			t.Fatalf("expected one balls got %d", len(got))
		}

		diff := cmp.Diff(got[0], seed[1], cmpopts.EquateApproxTime(time.Second), cmpopts.IgnoreFields(Ball{}, "ID"))
		if diff != "" {
			t.Fatalf("(-got, +want):\n%s", diff)
		}
		if got[0].ID == 0 {
			t.Fatalf("expected db to set id")
		}
	})

	t.Run("success approval date filter", func(t *testing.T) {
		t.Parallel()

		db, cleanup := crdb.StartTestDB(t, false)
		t.Cleanup(cleanup)

		now := time.Now()
		seed := []Ball{
			{
				Brand: Hammer,
				Name:  "Black Widow Mania",
				ImageURL: &url.URL{
					Scheme: "http",
					Host:   "some-url",
				},
				ApprovalDate: now.AddDate(0, -4, 0),
			},
			{
				Brand: Ebonite,
				Name:  "The One Reverb",
				ImageURL: &url.URL{
					Scheme: "http",
					Host:   "some-url",
				},
				ApprovalDate: now,
			},
		}

		ctx := context.Background()

		for _, b := range seed {
			stmt := `INSERT INTO balls (brand, name, image_url, approved_at) VALUES ($1, $2, $3, $4)`
			_, err := db.Exec(ctx, stmt, b.Brand, b.Name, b.ImageURL, b.ApprovalDate)
			if err != nil {
				t.Fatal(err)
			}
		}

		s := NewCRDBStore(db)

		got, err := s.GetAllBalls(ctx, BallFilter{
			ApprovalDate: &seed[0].ApprovalDate,
		})
		if err != nil {
			t.Fatal("expected error but got nil")
		}

		if len(got) != 1 {
			t.Fatalf("expected two balls got %d", len(got))
		}

		diff := cmp.Diff(got[0], seed[0], cmpopts.EquateApproxTime(time.Second), cmpopts.IgnoreFields(Ball{}, "ID"))
		if diff != "" {
			t.Fatalf("(-got, +want):\n%s", diff)
		}
		if got[0].ID == 0 {
			t.Fatalf("expected db to set id")
		}
	})

	t.Run("success all filters", func(t *testing.T) {
		t.Parallel()

		db, cleanup := crdb.StartTestDB(t, false)
		t.Cleanup(cleanup)

		now := time.Now()
		seed := []Ball{
			{
				Brand: Ebonite,
				Name:  "Turbo X",
				ImageURL: &url.URL{
					Scheme: "http",
					Host:   "some-url",
				},
				ApprovalDate: now.AddDate(-20, 0, 0),
			},
			{
				Brand: Ebonite,
				Name:  "Turbo X",
				ImageURL: &url.URL{
					Scheme: "http",
					Host:   "some-url",
				},
				ApprovalDate: now,
			},
		}

		ctx := context.Background()

		for _, b := range seed {
			stmt := `INSERT INTO balls (brand, name, image_url, approved_at) VALUES ($1, $2, $3, $4)`
			_, err := db.Exec(ctx, stmt, b.Brand, b.Name, b.ImageURL, b.ApprovalDate)
			if err != nil {
				t.Fatal(err)
			}
		}

		s := NewCRDBStore(db)

		got, err := s.GetAllBalls(ctx, BallFilter{
			Brand:        &seed[0].Brand,
			Name:         &seed[0].Name,
			ApprovalDate: &seed[0].ApprovalDate,
		})
		if err != nil {
			t.Fatal("expected error but got nil")
		}

		if len(got) != 1 {
			t.Fatalf("expected one balls got %d", len(got))
		}

		diff := cmp.Diff(got[0], seed[0], cmpopts.EquateApproxTime(time.Second), cmpopts.IgnoreFields(Ball{}, "ID"))
		if diff != "" {
			t.Fatalf("(-got, +want):\n%s", diff)
		}
		if got[0].ID == 0 {
			t.Fatalf("expected db to set id")
		}
	})
}
