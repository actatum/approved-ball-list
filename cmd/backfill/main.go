// Package main is the entrypoint for the backfill utility.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/actatum/approved-ball-list/internal/balls"
	"github.com/actatum/approved-ball-list/internal/crdb"
	"github.com/actatum/approved-ball-list/internal/mocks"
	"github.com/actatum/approved-ball-list/internal/usbc"
	"github.com/rs/zerolog"

	// imported for side effects
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}

// config holds the configurable values for the service.
type config struct {
	Env            string
	CockroachDBURL string
	Timeout        time.Duration
}

func newConfig() config {
	var cfg config
	var timeoutStr string
	flag.StringVar(&cfg.Env, "env", lookupEnv("ENV", "local"), "environment service is running in")
	flag.StringVar(&cfg.CockroachDBURL, "crdb-url", lookupEnv("COCKROACHDB_URL", ""), "cockroachdb url")
	flag.StringVar(&timeoutStr, "timeout", lookupEnv("TIMEOUT", "1m"), "timeout")
	flag.Parse()

	var err error
	cfg.Timeout, err = time.ParseDuration(timeoutStr)
	if err != nil {
		panic(err)
	}

	return cfg
}

// Run backfills the database but does not send notifications.
func Run() error {
	cfg := newConfig()

	zerolog.TimeFieldFormat = time.RFC3339

	var logger zerolog.Logger
	{
		if cfg.Env == "local" {
			logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
				Level(zerolog.TraceLevel).With().Timestamp().Logger()
		} else {
			logger = zerolog.New(os.Stdout).
				Level(zerolog.TraceLevel).With().Timestamp().Logger()
		}
	}

	var db *sql.DB
	{
		var err error
		db, err = sql.Open("pgx", cfg.CockroachDBURL)
		if err != nil {
			return fmt.Errorf("connecting to database: %w", err)
		}
		defer db.Close()
	}

	var repo balls.Repository
	{
		var err error
		repo, err = crdb.NewRepository(db)
		if err != nil {
			return fmt.Errorf("NewRepository: %w", err)
		}
	}

	var notificationService balls.NotificationService
	{
		notificationService = &mocks.NotificationServiceMock{
			SendNotificationFunc: func(_ context.Context, _ []balls.Ball) error {
				return nil
			},
		}
	}

	var usbcClient balls.USBCService
	{
		client := usbc.NewClient(&usbc.Config{
			Logger:     nil,
			HTTPClient: &http.Client{},
		})
		defer client.Close()

		usbcClient = client
	}

	svc := balls.NewService(repo, usbcClient, notificationService)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	ctx = logger.WithContext(ctx)

	start := time.Now()
	err := svc.CheckForNewlyApprovedBalls(ctx)
	if err != nil {
		return fmt.Errorf("checking for newly approved balls: %w", err)
	}
	logger.Info().Msgf("finished backfill in %s", time.Since(start))

	return nil
}

func lookupEnv(key string, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultValue
}
