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

// func TestMain(m *testing.M) {
// 	// command to start firestore emulator
// 	cmd := exec.Command("gcloud", "beta", "emulators", "firestore", "start", "--host-port=localhost", "--quiet")

// 	// this makes it killable
// 	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

// 	// we need to capture it's output to know when it's started
// 	stderr, err := cmd.StderrPipe()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer func() {
// 		closeErr := stderr.Close()
// 		if closeErr != nil {
// 			log.Println(closeErr)
// 		}
// 	}()

// 	// start her up!
// 	if err = cmd.Start(); err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println("firestore emulator pid", -cmd.Process.Pid)

// 	var result int
// 	defer func() {
// 		killErr := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
// 		if killErr != nil {
// 			log.Println(killErr)
// 		}
// 		os.Exit(result)
// 	}()

// 	// we're going to wait until it's running to start
// 	var wg sync.WaitGroup
// 	wg.Add(1)

// 	// by starting a separate go routine
// 	go func() {
// 		// reading it's output
// 		buf := make([]byte, 256)
// 		for {
// 			var n int
// 			n, err = stderr.Read(buf[:])
// 			if err != nil {
// 				// until it ends
// 				if err == io.EOF {
// 					break
// 				}
// 				// log.Fatalf("reading stderr %v", err)
// 			}

// 			if n > 0 {
// 				d := string(buf[:n])

// 				// only required if we want to see the emulator output
// 				log.Printf("%s", d)

// 				// checking for the message that it's started
// 				if strings.Contains(d, "Dev App Server is now running") {
// 					wg.Done()
// 				}

// 				// and capturing the FIRESTORE_EMULATOR_HOST value to set
// 				pos := strings.Index(d, firestoreEmulatorHost+"=")
// 				if pos > 0 {
// 					host := d[pos+len(firestoreEmulatorHost)+1 : len(d)-1]
// 					if err = os.Setenv(firestoreEmulatorHost, host); err != nil {
// 						log.Fatal(err)
// 					}
// 				}
// 			}
// 		}
// 	}()

// 	// wait until the running message has been received
// 	wg.Wait()
// 	result = m.Run()
// }

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
