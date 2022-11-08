// Package sqlite provides an implementation of the Repository using sqlite as the backing store.
package sqlite

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/actatum/approved-ball-list/internal/abl"
	"github.com/jmoiron/sqlx"

	// imported for side effects
	_ "modernc.org/sqlite"
)

// YYYYMMDD is the storage format for approval dates.
const YYYYMMDD = "2006-01-02"

// Repository handles storing data in sqlite.
type Repository struct {
	db *sqlx.DB
	sb sq.StatementBuilderType

	backupManager BackupManager

	file string

	io.Closer
}

// NewRepository returns a new instance of Repository.
func NewRepository(url string, backupManager BackupManager) (*Repository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if url != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(url), 0700); err != nil {
			return nil, err
		}

		err := backupManager.Restore(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("backupManager.Restore: %w", err)
		}
	}
	filename := url
	url += "?_journal=WAL&_timeout=5000&_fk=true"

	db, err := sqlx.ConnectContext(ctx, "sqlite", url)
	if err != nil {
		return nil, err
	}

	if err = runMigrations(db.DB); err != nil {
		return nil, err
	}

	return &Repository{
		db:            db,
		sb:            sq.StatementBuilder.PlaceholderFormat(sq.Question),
		backupManager: backupManager,
		file:          filename,
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
			Brand:        b.Brand,
			Name:         b.Name,
			ApprovalDate: b.ApprovedAt.Format(YYYYMMDD),
			ImageURL:     b.ImageURL,
		})
	}

	return withTx(ctx, r.db, func(tx *sqlx.Tx) error {
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
	err = withTx(ctx, r.db, func(tx *sqlx.Tx) error {
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
		approvedAt, _ := time.Parse(YYYYMMDD, b.ApprovalDate)
		balls = append(balls, abl.Ball{
			ID:         b.ID,
			Brand:      b.Brand,
			Name:       b.Name,
			ApprovedAt: approvedAt,
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
			Brand:        b.Brand,
			Name:         b.Name,
			ApprovalDate: b.ApprovedAt.Format(YYYYMMDD),
			ImageURL:     b.ImageURL,
		})
	}

	return withTx(ctx, r.db, func(tx *sqlx.Tx) error {
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

// Close shuts down the sql db and then backs up the file using the backup manager.
func (r *Repository) Close() error {
	if err := r.db.Close(); err != nil {
		return err
	}

	if r.file == ":memory:" {
		return nil
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer func(start time.Time) {
		fmt.Println(time.Since(start))
	}(start)

	return r.backupManager.Backup(ctx, r.file)
}

func withTx(ctx context.Context, db *sqlx.DB, txFn func(tx *sqlx.Tx) error) (err error) {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			// something went wrong, rollback
			_ = tx.Rollback()
		} else {
			// all good, commit
			err = tx.Commit()
		}
	}()

	err = txFn(tx)
	return err
}
