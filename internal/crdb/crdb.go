package crdb

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
)

//go:embed migrations/*.sql
var migrations embed.FS

const migrationVersion = 3

// NewDB returns a new pgxpool with the migrations applied to the database.
func NewDB(dsn string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	if err = runMigrations(stdlib.OpenDBFromPool(db)); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return errors.Wrap(err, "iofs.New")
	}

	driver, err := cockroachdb.WithInstance(db, &cockroachdb.Config{})
	if err != nil {
		return errors.Wrap(err, "cockroachdb.WithInstance")
	}

	m, err := migrate.NewWithInstance("iofs", source, "cockroachdb", driver)
	if err != nil {
		return errors.New("migrate.NewWithInstance")
	}

	err = m.Migrate(migrationVersion)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return source.Close()
}
