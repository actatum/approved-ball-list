package balls

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	AddBalls(ctx context.Context, balls []Ball) error
	GetAllBalls(ctx context.Context, filter BallFilter) ([]Ball, error)
}

type CRDBStore struct {
	db *pgxpool.Pool
}

func NewCRDBStore(db *pgxpool.Pool) *CRDBStore {
	return &CRDBStore{db: db}
}

func (s *CRDBStore) AddBalls(ctx context.Context, balls []Ball) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, ball := range balls {
		args := pgx.NamedArgs{
			"brand":       ball.Brand,
			"name":        ball.Name,
			"image_url":   ball.ImageURL,
			"approved_at": ball.ApprovalDate,
		}

		stmt := `
		INSERT INTO balls (brand, name, image_url, approved_at) VALUES (@brand, @name, @image_url, @approved_at)
		`

		if _, err = tx.Exec(ctx, stmt, args); err != nil {
			return fmt.Errorf("exec: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (s *CRDBStore) GetAllBalls(ctx context.Context, filter BallFilter) ([]Ball, error) {
	where, args := []string{"1 = 1"}, pgx.NamedArgs{}
	if filter.Brand != nil {
		where = append(where, "brand = @brand")
		args["brand"] = *filter.Brand
	}
	if filter.Name != nil {
		where = append(where, "name = @name")
		args["name"] = *filter.Name
	}
	if filter.ApprovalDate != nil {
		where = append(where, "approved_at = @approved_at")
		args["approved_at"] = *filter.ApprovalDate
	}

	stmt := `
	SELECT
		id,
		brand,
		name,
		approved_at,
		image_url
	FROM balls
	WHERE ` + strings.Join(where, " AND ")
	rows, err := s.db.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var balls []Ball
	for rows.Next() {
		var ball Ball
		var imageURL string
		err = rows.Scan(
			&ball.ID,
			&ball.Brand,
			&ball.Name,
			&ball.ApprovalDate,
			&imageURL,
		)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		ball.ImageURL, err = url.Parse(imageURL)
		if err != nil {
			return nil, fmt.Errorf("parsing image url: %w", err)
		}

		balls = append(balls, ball)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return balls, nil
}
