// Package crdb provides an implementation of the Repository using cockroachdb as the backing store.
package crdb

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/actatum/approved-ball-list/internal/balls"
	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/rs/zerolog"

	// imported for side effects
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Repository handles storing data in sqlite.
type Repository struct {
	sb sq.StatementBuilderType
	db *sql.DB
}

// NewRepository returns a new instance of Repository.
func NewRepository(db *sql.DB) (*Repository, error) {
	if db == nil {
		return nil, fmt.Errorf("nil db")
	}

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return &Repository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

// Add adds balls to the db if they don't already exist.
func (r *Repository) Add(ctx context.Context, items ...balls.Ball) ([]balls.Ball, error) {
	if len(items) == 0 {
		return nil, nil
	}

	var added []balls.Ball
	err := crdb.ExecuteTx(ctx, r.db, &sql.TxOptions{}, func(tx *sql.Tx) error {
		for _, b := range items {
			row := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM balls WHERE brand = $1 AND name = $2`, b.Brand, b.Name)
			var cnt int
			if err := row.Scan(&cnt); err != nil {
				return fmt.Errorf("scanning count: %w", err)
			}

			if cnt == 0 {
				zerolog.Ctx(ctx).Info().Msgf("adding ball %s %s", b.Brand, b.Name)
				_, err := tx.ExecContext(
					ctx,
					`INSERT INTO balls (brand, name, image_url, approved_at) VALUES ($1, $2, $3, $4)`,
					b.Brand, b.Name, b.ImageURL, b.ApprovalDate,
				)
				if err != nil {
					return fmt.Errorf("inserting row: %w", err)
				}

				added = append(added, b)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return added, nil
}
