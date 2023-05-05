// Package crdb provides an implementation of the Repository using cockroachdb as the backing store.
package crdb

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/actatum/approved-ball-list/internal/abl"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"

	// imported for side effects
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Repository handles storing data in sqlite.
type Repository struct {
	db *sqlx.DB
	sb sq.StatementBuilderType
}

// NewRepository returns a new instance of Repository.
func NewRepository(db *sqlx.DB) (*Repository, error) {
	if err := runMigrations(db.DB); err != nil {
		return nil, err
	}

	return &Repository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

// AddBalls stores the given balls in the repository.
func (r *Repository) AddBalls(ctx context.Context, balls []abl.Ball) error {
	if len(balls) == 0 {
		return nil
	}

	batch := make([]ball, 0, len(balls))
	for _, b := range balls {
		batch = append(batch, ball{
			Brand:      b.Brand,
			Name:       b.Name,
			ApprovedAt: b.ApprovedAt,
			ImageURL:   b.ImageURL,
		})
	}

	return crdbsqlx.ExecuteTx(ctx, r.db, &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		_, err := tx.NamedExec(insertBallsQuery, batch)
		if err != nil {
			return fmt.Errorf("tx.NamedExec: %w", err)
		}

		return nil
	})
}

// ListBalls returns a list of balls based on the given filter.
func (r *Repository) ListBalls(ctx context.Context, filter abl.BallFilter) (abl.ListBallsResult, error) {
	query, args, offset, err := listBallsQuery(&r.sb, filter)
	if err != nil {
		return abl.ListBallsResult{}, fmt.Errorf("listBallsQuery: %w", err)
	}

	var dbBalls []listBallRow
	err = crdbsqlx.ExecuteTx(ctx, r.db, &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		err := tx.Select(&dbBalls, query, args...)
		if err != nil {
			return fmt.Errorf("tx.Select: %w", err)
		}

		return nil
	})
	if err != nil {
		return abl.ListBallsResult{}, err
	}

	balls := make([]abl.Ball, 0, len(dbBalls))
	for _, b := range dbBalls {
		balls = append(balls, abl.Ball{
			ID:         strconv.FormatInt(b.ID, 10),
			Brand:      b.Brand,
			Name:       b.Name,
			ApprovedAt: b.ApprovedAt,
			ImageURL:   b.ImageURL,
		})
	}

	var npt string
	if len(balls) > filter.PageSize {
		newOffset := offset + filter.PageSize
		npt = base64.URLEncoding.EncodeToString([]byte(strconv.Itoa(newOffset)))
		balls = balls[:len(balls)-1]
	}

	var cnt int
	if len(dbBalls) > 0 {
		cnt = dbBalls[0].Count
	}

	return abl.ListBallsResult{
		Balls:         balls,
		NextPageToken: npt,
		Count:         cnt,
	}, nil
}

// RemoveBalls removes the given balls from the database.
func (r *Repository) RemoveBalls(ctx context.Context, balls []abl.Ball) error {
	if len(balls) == 0 {
		return nil
	}

	batch := make([]ball, 0, len(balls))
	for _, b := range balls {
		batch = append(batch, ball{
			Brand:      b.Brand,
			Name:       b.Name,
			ApprovedAt: b.ApprovedAt,
			ImageURL:   b.ImageURL,
		})
	}

	return crdbsqlx.ExecuteTx(ctx, r.db, &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		query := `DELETE FROM balls WHERE (brand, name) IN (`
		for idx, b := range batch {
			tuple := fmt.Sprintf("('%s','%s')", b.Brand, b.Name)
			if idx != len(batch)-1 {
				tuple += ","
			}
			query += tuple
		}
		query += ")"

		_, err := tx.Exec(query)
		if err != nil {
			return fmt.Errorf("tx.Exec: %w", err)
		}

		return nil
	})
}
