// Package main provides the entrypoint to the service when its run as a server.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/actatum/approved-ball-list/internal/balls"
	"github.com/actatum/approved-ball-list/internal/crdb"
	"github.com/actatum/approved-ball-list/internal/log"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var (
		cockroachURL             = flag.String("crdb-url", lookupEnv("COCKROACHDB_URL", ""), "cockroachdb url")
		discordChannels channels = strings.Split(lookupEnv("DISCORD_CHANNELS", ""), ",")
		discordToken             = flag.String("discord-token", lookupEnv("DISCORD_TOKEN", ""), "discord bot token")
		env                      = flag.String("env", lookupEnv("ENV", "local"), "environment service is running in")
		port                     = flag.String("port", lookupEnv("PORT", "8080"), "http server port")
	)
	flag.Var(&discordChannels, "discord-channels", "discord channels to notify")
	flag.Parse()

	logger := log.NewLogger(os.Stderr)

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

	var notifier balls.Notifier
	{
		switch *env {
		case "prod":
			dg, err := discordgo.New(fmt.Sprintf("Bot %s", *discordToken))
			if err != nil {
				logger.Error("error creating discord client", slog.Any("error", err))
				os.Exit(1)
			}
			defer dg.Close()

			notifier = balls.NewDiscordNotifier(dg, discordChannels)

		default:
			notifier = balls.LocalNotifier{}
		}
	}

	usbcService := balls.NewHTTPUSBCService(&http.Client{}, logger)
	service := balls.NewService(logger, store, usbcService, notifier)

	h := balls.NewHTTPHandler(logger, service, *env)

	errs := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		srv := &http.Server{
			Addr:         fmt.Sprintf(":%s", *port),
			Handler:      h,
			IdleTimeout:  time.Minute,
			ReadTimeout:  time.Minute,
			WriteTimeout: time.Minute,
		}
		logger.Info("starting http server", slog.String("addr", srv.Addr))
		errs <- srv.ListenAndServe()
	}()

	logger.Info("shutting down", slog.Any("exit", <-errs))
}

func lookupEnv(key string, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultValue
}

type channels []string

func (c *channels) String() string {
	return strings.Join(*c, ",")
}

func (c *channels) Set(value string) error {
	*c = strings.Split(value, ",")
	return nil
}
