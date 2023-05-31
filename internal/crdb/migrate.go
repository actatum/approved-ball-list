package crdb

import (
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/pkg/errors"
)

//go:embed migrations/*.sql
var migrations embed.FS

const migrationVersion = 2

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
