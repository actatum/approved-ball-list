package crdb

import (
	"log"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// StartTestDB starts a cockroachdb container for use in testing.
func StartTestDB(tb testing.TB, logsEnabled bool) (*pgxpool.Pool, func()) {
	tb.Helper()

	crdbURL := &url.URL{
		Scheme: "postgresql",
		User:   url.User("root"),
		Path:   "/defaultdb",
	}
	q := crdbURL.Query()
	q.Add("sslmode", "disable")
	crdbURL.RawQuery = q.Encode()

	pool, err := dockertest.NewPool("")
	if err != nil {
		tb.Fatalf("could not connect to docker: %v", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "cockroachdb/cockroach",
		Tag:        "latest",
		Cmd:        []string{"start-single-node", "--insecure"},
	}, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.NeverRestart()
	})
	if err != nil {
		tb.Fatalf("could not start cockroachdb container: %v", err)
	}

	crdbURL.Host = getHostPort(resource, "26257/tcp")

	var logWaiter docker.CloseWaiter
	if logsEnabled {
		logWaiter, err = pool.Client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
			Container:    resource.Container.ID,
			OutputStream: log.Writer(),
			ErrorStream:  log.Writer(),
			Stderr:       true,
			Stdout:       true,
			Stream:       true,
		})
		if err != nil {
			tb.Fatalf("could not connect to cockroachdb container log output: %v", err)
		}
	}

	pool.MaxWait = 15 * time.Second
	var db *pgxpool.Pool
	err = pool.Retry(func() error {
		db, err = NewDB(crdbURL.String())
		return err
	})
	if err != nil {
		tb.Fatalf("could not connect to cockroachdb container: %v", err)
	}

	return db, func() {
		db.Close()

		if err := pool.Purge(resource); err != nil {
			tb.Errorf("pool.Purge(resource) error = %v", err)
		}
		if logWaiter != nil {
			if err := logWaiter.Close(); err != nil {
				tb.Errorf("logWaiter.Close() error = %v", err)
			}
		}
	}
}

func getHostPort(resource *dockertest.Resource, id string) string {
	dockerURL := os.Getenv("DOCKER_HOST")
	if dockerURL == "" {
		return resource.GetHostPort(id)
	}
	u, err := url.Parse(dockerURL)
	if err != nil {
		panic(err)
	}
	return u.Hostname() + ":" + resource.GetPort(id)
}
