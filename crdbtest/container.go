package crdbtest

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func StartTestContainer(tb testing.TB) (db *sql.DB, close func()) {
	tb.Helper()

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "cockroachdb/cockroach:v23.1.13",
		ExposedPorts: []string{"26257/tcp", "8080/tcp"},
		WaitingFor:   wait.ForHTTP("/health").WithPort("8080"),
		Cmd:          []string{"start-single-node", "--insecure"},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		tb.Fatal(err)
	}

	mappedPort, err := container.MappedPort(ctx, "26257")
	if err != nil {
		tb.Fatal(err)
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		tb.Fatal(err)
	}

	uri := fmt.Sprintf("postgres://root@%s:%s", hostIP, mappedPort.Port())

	db, err = sql.Open("pgx", uri)
	if err != nil {
		tb.Fatal(err)
	}

	dir := os.DirFS("../migrations")

	files, err := fs.Glob(dir, "*.sql")
	if err != nil {
		tb.Fatal(err)
	}
	sort.Strings(files)

	for _, file := range files {
		buf, err := fs.ReadFile(dir, file)
		if err != nil {
			tb.Fatal(err)
		}

		time.Sleep(50 * time.Millisecond)
		if err := migrateFile(tb, ctx, db, buf); err != nil {
			tb.Fatal(err)
		}
	}

	return db, func() {
		err = container.Terminate(ctx)
		if err != nil {
			tb.Fatal(err)
		}
	}
}

func migrateFile(tb testing.TB, ctx context.Context, db *sql.DB, file []byte) error {
	tb.Helper()

	return crdb.ExecuteTx(ctx, db, &sql.TxOptions{}, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, string(file)); err != nil {
			return err
		}

		return nil
	})
}
