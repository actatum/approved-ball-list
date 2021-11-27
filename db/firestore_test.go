package db

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/actatum/approved-ball-list/models"
	"github.com/stretchr/testify/assert"
)

const testProjectID = "test-project-id"
const firestoreEmulatorHost = "FIRESTORE_EMULATOR_HOST"

func TestMain(m *testing.M) {
	// command to start firestore emulator
	cmd := exec.Command("gcloud", "beta", "emulators", "firestore", "start", "--host-port=localhost")

	// this makes it killable
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// we need to capture it's output to know when it's started
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	defer stderr.Close()

	// start her up!
	if err = cmd.Start(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("firestore emulator pid", -cmd.Process.Pid)

	var result int
	defer func() {
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		os.Exit(result)
	}()

	// we're going to wait until it's running to start
	var wg sync.WaitGroup
	wg.Add(1)

	// by starting a separate go routine
	go func() {
		// reading it's output
		buf := make([]byte, 256, 256)
		for {
			n, err := stderr.Read(buf[:])
			if err != nil {
				// until it ends
				if err == io.EOF {
					break
				}
				// log.Fatalf("reading stderr %v", err)
			}

			if n > 0 {
				d := string(buf[:n])

				// only required if we want to see the emulator output
				log.Printf("%s", d)

				// checking for the message that it's started
				if strings.Contains(d, "Dev App Server is now running") {
					wg.Done()
				}

				// and capturing the FIRESTORE_EMULATOR_HOST value to set
				pos := strings.Index(d, firestoreEmulatorHost+"=")
				if pos > 0 {
					host := d[pos+len(firestoreEmulatorHost)+1 : len(d)-1]
					os.Setenv(firestoreEmulatorHost, host)
				}
			}
		}
	}()

	// wait until the running message has been received
	wg.Wait()
	result = m.Run()
}

func TestAddGetAndClearBalls(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := firestore.NewClient(ctx, testProjectID)
	if err != nil {
		t.Fatalf("error creating firestore client: %v", err)
	}
	t.Cleanup(func() {
		cerr := client.Close()
		if cerr != nil {
			fmt.Println(cerr)
		}
		cancel()
	})

	numBalls := 501
	balls := genBalls(numBalls)
	err = AddBalls(ctx, client, balls)
	if err != nil {
		t.Fatalf("error adding balls to database: %v", err)
	}

	fromDB, err := GetAllBalls(ctx, client)
	if err != nil {
		t.Fatalf("error retreiving balls from database: %v", err)
	}

	assert.Equal(t, numBalls, len(fromDB))

	err = ClearCollection(ctx, client)
	if err != nil {
		t.Fatalf("error clearing collection: %v", err)
	}

	fromDB, err = GetAllBalls(ctx, client)
	if err != nil {
		t.Fatalf("error retreiving balls from database: %v", err)
	}

	assert.Equal(t, 0, len(fromDB))
}

func genBalls(n int) []models.Ball {
	strLen := 10
	rand.Seed(time.Now().UnixNano())
	balls := make([]models.Ball, n)
	for i := 0; i < n; i++ {
		b := models.Ball{
			XMLName:      xml.Name{},
			Brand:        randSeq(strLen),
			Name:         randSeq(strLen),
			DateApproved: randSeq(strLen),
			ImageURL:     randSeq(strLen),
		}
		balls[i] = b
	}

	return balls
}

func randSeq(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
