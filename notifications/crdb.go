package notifications

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type CRDBRepository struct {
	db *sql.DB
}

func NewCRDBRepository(db *sql.DB) *CRDBRepository {
	return &CRDBRepository{
		db: db,
	}
}

func (r *CRDBRepository) StoreTarget(ctx context.Context, target Target) error {
	return crdb.ExecuteTx(ctx, r.db, &sql.TxOptions{}, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO notification_targets (id, created_at, updated_at, type, destination) 
		VALUES ($1, $2, $3, $4, $5)`,
			target.ID, target.CreatedAt, target.UpdateAt, target.Type, target.Destination,
		)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == pgerrcode.UniqueViolation && strings.Contains(pgErr.Message, "unique_destination") {
					return DuplicateTargetError{targetType: target.Type, destination: target.Destination}
				}
			}

			return err
		}

		return nil
	})
}

func (r *CRDBRepository) FindAllTargets(ctx context.Context) ([]Target, error) {
	targets := make([]Target, 0)
	err := crdb.ExecuteTx(ctx, r.db, &sql.TxOptions{}, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `SELECT id, created_at, updated_at, type, destination FROM notification_targets`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var i Target
			if err := rows.Scan(&i.ID, &i.CreatedAt, &i.UpdateAt, &i.Type, &i.Destination); err != nil {
				return err
			}
			targets = append(targets, i)
		}
		if err := rows.Close(); err != nil {
			return err
		}
		if err := rows.Err(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return targets, nil
}

func (r *CRDBRepository) Store(ctx context.Context, notifications []Notification) error {
	return crdb.ExecuteTx(ctx, r.db, &sql.TxOptions{}, func(tx *sql.Tx) error {
		for _, notif := range notifications {
			ballID, err := strconv.ParseInt(notif.Content[0].ID, 10, 64)
			if err != nil {
				return err
			}

			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO notifications (id, state, ball_id, target_id) VALUES ($1, $2, $3, $4)`,
				notif.ID, notif.State, ballID, notif.Target.ID,
			)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
