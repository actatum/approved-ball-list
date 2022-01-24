package repository

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/actatum/approved-ball-list/core"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
)

const testProjectID = "test-project-id"
const firestoreEmulatorHost = "FIRESTORE_EMULATOR_HOST"

func startDatabase(tb testing.TB) string {
	tb.Helper()

	pool, err := dockertest.NewPool("")
	if err != nil {
		tb.Fatalf("Could not connect to docker: %v", err)
	}

	env := []string{
		"FIRESTORE_PROJECT_ID=" + testProjectID,
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mtlynch/firestore-emulator-docker",
		Tag:        "latest",
		Env:        env,
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		tb.Fatalf("Could not start firestore container: %v", err)
	}
	tb.Cleanup(func() {
		err = pool.Purge(resource)
		if err != nil {
			tb.Fatalf("Could not purge container: %v", err)
		}
	})

	pubsubURL := getHostPort(resource, "8080/tcp")
	fmt.Println(pubsubURL)

	logWaiter, err := pool.Client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    resource.Container.ID,
		OutputStream: log.Writer(),
		ErrorStream:  log.Writer(),
		Stderr:       true,
		Stdout:       true,
		Stream:       true,
	})
	if err != nil {
		tb.Fatalf("Could not connect to firestore container log output: %v", err)
	}

	tb.Cleanup(func() {
		err = logWaiter.Close()
		if err != nil {
			tb.Fatalf("Could not close container log: %v", err)
		}
		err = logWaiter.Wait()
		if err != nil {
			tb.Fatalf("Could not wait for container log to close: %v", err)
		}
	})

	tb.Setenv(firestoreEmulatorHost, pubsubURL)

	pool.MaxWait = 10 * time.Second
	err = pool.Retry(func() (err error) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		client, err := firestore.NewClient(ctx, testProjectID)
		if err != nil {
			return err
		}

		defer func() {
			cerr := client.Close()
			if err == nil {
				err = cerr
			}
		}()

		_, _, err = client.Collection("random").Add(ctx, map[string]string{"hello": "world"})

		return err
	})
	if err != nil {
		tb.Fatalf("Could not connect to firestore container: %v", err)
	}

	return pubsubURL
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

func TestRepository_InsertNewBalls(t *testing.T) {
	type args struct {
		ctx   context.Context
		balls []core.Ball
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "insert 0 balls",
			args: args{
				ctx:   context.Background(),
				balls: nil,
			},
			wantErr: false,
		},
		{
			name: "insert a few balls",
			args: args{
				ctx: context.Background(),
				balls: []core.Ball{
					{
						Brand: "Storm",
						Name:  "Nova",
					},
					{
						Brand: "Storm",
						Name:  "Spectre",
					},
					{
						Brand: "Roto Grip",
						Name:  "Hustle Solid",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVal := startDatabase(t)
			t.Setenv(firestoreEmulatorHost, envVal)

			r, err := NewRepository(tt.args.ctx, testProjectID)
			if err != nil {
				t.Fatalf("failed to initialize repository: %v", err)
			}

			t.Cleanup(func() {
				clearErr := r.ClearCollection(tt.args.ctx)
				if clearErr != nil {
					t.Fatalf("failed to clear collection during cleanup: %v", clearErr)
				}

				closeErr := r.Close()
				if closeErr != nil {
					t.Fatalf("failed to close repo: %v", closeErr)
				}
			})

			if err := r.InsertNewBalls(tt.args.ctx, tt.args.balls); (err != nil) != tt.wantErr {
				t.Errorf("Repository.InsertNewBalls() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepository_GetAllBalls(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    []core.Ball
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "get 0 balls",
			args: args{
				ctx: context.Background(),
			},
			want:    []core.Ball{},
			wantErr: false,
		},
		{
			name: "get a few balls",
			args: args{
				ctx: context.Background(),
			},
			want: []core.Ball{
				{
					Brand: "Storm",
					Name:  "Nova",
				},
				{
					Brand: "Storm",
					Name:  "Spectre",
				},
				{
					Brand: "Roto Grip",
					Name:  "Hustle Solid",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVal := startDatabase(t)
			t.Setenv(firestoreEmulatorHost, envVal)
			r, err := NewRepository(tt.args.ctx, testProjectID)
			if err != nil {
				t.Fatalf("failed to initialize repository: %v", err)
			}

			t.Cleanup(func() {
				clearErr := r.ClearCollection(tt.args.ctx)
				if clearErr != nil {
					t.Fatalf("failed to clear collection during cleanup: %v", clearErr)
				}

				closeErr := r.Close()
				if closeErr != nil {
					t.Fatalf("failed to close repo: %v", closeErr)
				}
			})

			if err = r.InsertNewBalls(tt.args.ctx, tt.want); err != nil {
				t.Fatalf("error inserting balls: %v", err)
			}

			got, err := r.GetAllBalls(tt.args.ctx)

			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			for _, b := range tt.want {
				assert.True(t, contains(got, b))
			}
		})
	}
}

func contains(s []core.Ball, e core.Ball) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
