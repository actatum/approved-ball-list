package cockroachdb

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"

	// imported for side effects
	_ "github.com/jackc/pgx/v4/stdlib"
)

//go:embed migrations/*.sql
var fs embed.FS

const version = 1

// NewDB returns a new sqlx.DB with the given dsn.
func NewDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlx.Connect: %v", err)
	}

	return db, validateSchema(db.DB)
}

func validateSchema(db *sql.DB) error {
	sourceInstance, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

	driverInstance, err := cockroachdb.WithInstance(db, new(cockroachdb.Config))
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", sourceInstance, "cockroachdb", driverInstance)
	if err != nil {
		return err
	}
	err = m.Migrate(version) // current version
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return sourceInstance.Close()
}
