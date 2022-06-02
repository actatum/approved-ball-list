package cockroachdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/actatum/approved-ball-list/abl"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
)

type ball struct {
	ID         int       `db:"id"`
	Brand      string    `db:"brand"`
	Name       string    `db:"name"`
	ApprovedAt time.Time `db:"approved_at"`
	ImageURL   string    `db:"image_url"`
}

// BallService represents a service for managing our copy of the USBC's approved ball list.
type BallService struct {
	db *sqlx.DB
	sb sq.StatementBuilderType
}

// NewBallService returns a new instance of BallService.
func NewBallService(db *sqlx.DB) *BallService {
	return &BallService{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// AddBalls adds a batch of balls.
func (s *BallService) AddBalls(ctx context.Context, balls ...abl.Ball) error {
	batch := make([]ball, 0, len(balls))
	for i := range balls {
		batch = append(batch, ball{
			Brand:      balls[i].Brand,
			Name:       balls[i].Name,
			ApprovedAt: balls[i].ApprovedAt,
			ImageURL:   balls[i].ImageURL,
		})
	}

	return crdbsqlx.ExecuteTx(ctx, s.db, &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		_, err := tx.NamedExecContext(
			ctx,
			`INSERT INTO balls (brand, name, image_url, approved_at) 
		VALUES (:brand, :name, :image_url, :approved_at)`,
			batch,
		)
		if err != nil {
			return err
		}

		return nil
	})
}

// ListBalls returns a list of all balls with the given filter.
func (s *BallService) ListBalls(ctx context.Context, filter abl.BallFilter) ([]abl.Ball, int, error) {
	balls := make([]abl.Ball, 0)
	cnt := 0

	err := crdbsqlx.ExecuteTx(ctx, s.db, &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		q := s.sb.
			Select("id", "brand", "name", "approved_at", "image_url", "COUNT(*) OVER()").
			From("balls")
		if filter.Brand != nil {
			q = q.Where(sq.Eq{"brand": *filter.Brand})
		}

		rows, err := q.RunWith(tx).QueryContext(ctx)
		if err != nil {
			return fmt.Errorf("tx.QueryContext: %v", err)
		}
		defer func() {
			_ = rows.Close()
		}()

		for rows.Next() {
			var b abl.Ball
			if err = rows.Scan(
				&b.ID,
				&b.Brand,
				&b.Name,
				&b.ApprovedAt,
				&b.ImageURL,
				&cnt,
			); err != nil {
				return fmt.Errorf("rows.Scan: %v", err)
			}

			balls = append(balls, b)
		}
		if err = rows.Err(); err != nil {
			return fmt.Errorf("rows.Err: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return balls, cnt, nil
}

// RemoveBalls removes a batch of balls.
func (s *BallService) RemoveBalls(ctx context.Context, balls ...abl.Ball) error {
	return crdbsqlx.ExecuteTx(ctx, s.db, &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		query := `DELETE FROM balls WHERE (brand, name) IN (`
		for i := range balls {
			tuple := fmt.Sprintf("('%s', '%s')", balls[i].Brand, balls[i].Name)
			if i != len(balls)-1 {
				tuple += ","
			}
			query += tuple
		}
		query += ")"

		_, err := tx.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("tx.ExecContext: %v", err)
		}

		return nil
	})
}
