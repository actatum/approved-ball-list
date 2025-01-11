// Package app provides types that coordinate all the pieces of the service into a runnable unit.
package app

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/actatum/approved-ball-list/internal/balls"
	"github.com/actatum/approved-ball-list/internal/crdb"
	"github.com/actatum/approved-ball-list/internal/discord"
	"github.com/actatum/approved-ball-list/internal/mocks"
	"github.com/actatum/approved-ball-list/internal/usbc"
	"github.com/oklog/run"
	"github.com/rs/zerolog"

	// imported for side effects
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Channels represents a slice of strings with discord channel id's.
type Channels []string

// String returns the string representation of the Channels type as a comma separated string.
func (c *Channels) String() string {
	return strings.Join(*c, ",")
}

// Set sets the value of the Channels slice.
func (c *Channels) Set(value string) error {
	*c = strings.Split(value, ",")
	return nil
}

// config holds the configurable values for the service.
type config struct {
	Env             string
	Port            string
	DiscordToken    string
	DiscordChannels Channels
	CockroachDBURL  string
}

func newConfig() config {
	var cfg config
	cfg.DiscordChannels = strings.Split(lookupEnv("DISCORD_CHANNELS", ""), ",")
	flag.StringVar(&cfg.Env, "env", lookupEnv("ENV", "local"), "environment service is running in")
	flag.StringVar(&cfg.Port, "port", lookupEnv("PORT", "8080"), "http server port")
	flag.StringVar(&cfg.DiscordToken, "discord-token", lookupEnv("DISCORD_TOKEN", ""), "discord bot token")
	flag.StringVar(&cfg.CockroachDBURL, "crdb-url", lookupEnv("COCKROACHDB_URL", ""), "cockroachdb url")
	flag.Var(&cfg.DiscordChannels, "discord-channels", "discord channels to notify")
	flag.Parse()

	return cfg
}

// Run runs the application.
func Run() error {
	cfg := newConfig()

	zerolog.TimeFieldFormat = time.RFC3339

	var logger zerolog.Logger
	{
		if cfg.Env == "local" {
			logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
				Level(zerolog.TraceLevel).With().Timestamp().Logger().Hook(&severityHook{})
		} else {
			logger = zerolog.New(os.Stdout).
				Level(zerolog.TraceLevel).With().Timestamp().Logger().Hook(&severityHook{})
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
		if cfg.Env != "local" {
			discordNotificationService, err := discord.NewNotifier(cfg.DiscordToken, cfg.DiscordChannels)
			if err != nil {
				return fmt.Errorf("NewNotifier: %w", err)
			}
			notificationService = discordNotificationService
			defer func() {
				e := discordNotificationService.Close()
				if e != nil {
					logger.Info().Err(e).Msg("Notifier.Close")
				}
			}()
		} else {
			notificationService = &mocks.NotificationServiceMock{
				SendNotificationFunc: func(_ context.Context, _ []balls.Ball) error {
					return nil
				},
			}
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

	h := &handler{
		svc:    svc,
		logger: &logger,
		cfg:    cfg,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", h.cfg.Port),
		Handler:      h.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	var g run.Group
	{
		// HTTP Server
		g.Add(func() error {
			h.logger.Info().Msgf("👋 HTTP server listening on :%s", h.cfg.Port)
			return srv.ListenAndServe()
		}, func(_ error) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if shutdownErr := srv.Shutdown(ctx); shutdownErr != nil {
				h.logger.Error().Err(shutdownErr).Msg("failed to shutdown HTTP server")
			}
		})
	}
	{
		// This function just sits and waits for ctrl-C.
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}

	return g.Run()
}

func lookupEnv(key string, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultValue
}
