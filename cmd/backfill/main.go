// Package main is the entrypoint for the backfill utility.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/actatum/approved-ball-list/internal/balls"
	"github.com/actatum/approved-ball-list/internal/crdb"
	"github.com/actatum/approved-ball-list/internal/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var (
		cockroachURL = flag.String("crdb-url", lookupEnv("COCKROACHDB_URL", ""), "cockroachdb url")
		timeout      = flag.Duration("timeout", 1*time.Minute, "max duration before process shuts down")
	)
	flag.Parse()

	logger := log.NewLogger(os.Stderr, log.WithFmtLog())

	var db *pgxpool.Pool
	{
		var err error
		db, err = crdb.NewDB(*cockroachURL)
		if err != nil {
			logger.Error("error connecting to cockroachdb", slog.Any("error", err))
			os.Exit(1)
		}
		defer db.Close()
	}

	store := balls.NewCRDBStore(db)
	notifier := balls.LocalNotifier{}
	usbcService := balls.NewHTTPUSBCService(&http.Client{}, logger)
	service := balls.NewService(logger, store, usbcService, notifier)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	start := time.Now()
	err := service.CheckForNewlyApprovedBalls(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "error checking for newly approved balls", slog.Any("error", err))
		os.Exit(1)
	}

	logger.InfoContext(ctx, fmt.Sprintf("finished backfill in %s", time.Since(start)))
}

func lookupEnv(key string, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultValue
}
