// Package app provides types that coordinate all the pieces of the service into a runnable unit.
package app

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/actatum/approved-ball-list/internal/abl"
	"github.com/actatum/approved-ball-list/internal/discord"
	"github.com/actatum/approved-ball-list/internal/mocks"
	"github.com/actatum/approved-ball-list/internal/sqlite"
	"github.com/actatum/approved-ball-list/internal/sqlite/gcs"
	"github.com/actatum/approved-ball-list/internal/usbc"
	"github.com/oklog/run"
	"github.com/rs/zerolog"
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
	StorageBucket   string
	DiscordToken    string
	DiscordChannels Channels
}

func newConfig() config {
	var cfg config
	cfg.DiscordChannels = strings.Split(lookupEnv("DISCORD_CHANNELS", ""), ",")
	flag.StringVar(&cfg.Env, "env", lookupEnv("ENV", "local"), "environment service is running in")
	flag.StringVar(&cfg.Port, "port", lookupEnv("PORT", "8080"), "http server port")
	flag.StringVar(&cfg.StorageBucket, "storage-bucket", lookupEnv("STORAGE_BUCKET", ""), "gcp storage bucket for backups")
	flag.StringVar(&cfg.DiscordToken, "discord-token", lookupEnv("DISCORD_TOKEN", ""), "discord bot token")
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

	var err error
	var bm sqlite.BackupManager
	{
		if cfg.Env != "local" {
			bm, err = gcs.NewBackupManager(cfg.StorageBucket)
			if err != nil {
				return fmt.Errorf("NewBackupManager: %w", err)
			}
		} else {
			bm = &mocks.BackupManagerMock{
				BackupFunc: func(ctx context.Context, file string) error {
					return nil
				},
				RestoreFunc: func(ctx context.Context, file string) error {
					return nil
				},
				CloseFunc: func() error {
					return nil
				},
			}
		}
	}
	defer func() {
		e := bm.Close()
		if e != nil {
			logger.Info().Err(e).Msg("BackupManager.Close")
		}
	}()

	var repo abl.Repository
	{
		repo, err = sqlite.NewRepository("sqlite.db", bm)
		if err != nil {
			return fmt.Errorf("NewRepository: %w", err)
		}
	}
	defer func() {
		e := repo.Close()
		if e != nil {
			logger.Info().Err(e).Msg("Repository.Close")
		}
	}()

	var notifier abl.Notifier
	{
		if cfg.Env != "local" {
			notifier, err = discord.NewNotifier(cfg.DiscordToken, cfg.DiscordChannels)
			if err != nil {
				return fmt.Errorf("NewNotifier: %w", err)
			}
		} else {
			notifier = &mocks.NotifierMock{
				NotifyFunc: func(ctx context.Context, notifications []abl.Notification) error {
					return nil
				},
				CloseFunc: func() error {
					return nil
				},
			}
		}
	}
	defer func() {
		e := notifier.Close()
		if e != nil {
			logger.Info().Err(e).Msg("Notifier.Close")
		}
	}()

	var usbcClient abl.USBCClient
	{
		usbcClient = usbc.NewClient(&usbc.Config{
			Logger:     nil,
			HTTPClient: &http.Client{},
		})
	}
	defer usbcClient.Close()

	svc := abl.NewService(repo, notifier, usbcClient)

	h := &handler{
		svc:    svc,
		logger: &logger,
		cfg:    cfg,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", h.cfg.Port),
		Handler:      h.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	var g run.Group
	{
		// HTTP Server
		g.Add(func() error {
			h.logger.Info().Msgf("👋 HTTP server listening on :%s", h.cfg.Port)
			return srv.ListenAndServe()
		}, func(err error) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				h.logger.Error().Err(err).Msg("failed to shutdown HTTP server")
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
