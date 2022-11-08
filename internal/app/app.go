// Package app provides types that coordinate all the pieces of the service into a runnable unit.
package app

import (
	"context"
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

// Config holds the configurable values for the service.
type Config struct {
	Env             string
	Port            string
	StorageBucket   string
	DiscordToken    string
	DiscordChannels Channels
}

// Application ...
type Application struct {
	config     Config
	service    abl.Service
	httpServer *http.Server
	logger     *zerolog.Logger
}

// NewApplication returns a new instance of the application.
func NewApplication(cfg Config) (*Application, error) {
	a := &Application{
		config: cfg,
	}
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
	a.logger = &logger

	logger.Info().Interface("config", cfg).Send()

	return a, nil
}

// Run runs the application.
func (a *Application) Run() error {
	var err error

	var bm sqlite.BackupManager
	{
		if a.config.Env != "local" {
			bm, err = gcs.NewBackupManager(a.config.StorageBucket)
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
			a.logger.Info().Err(e).Msg("BackupManager.Close")
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
			a.logger.Info().Err(e).Msg("Repository.Close")
		}
	}()

	var notifier abl.Notifier
	{
		if a.config.Env != "local" {
			notifier, err = discord.NewNotifier(a.config.DiscordToken, nil)
			if err != nil {
				return fmt.Errorf("NewNotifier")
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
			a.logger.Info().Err(e).Msg("Notifier.Close")
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
	a.service = svc

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", a.config.Port),
		Handler:      a.routes(a.logger),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	a.httpServer = srv

	var g run.Group
	{
		// HTTP Server
		g.Add(func() error {
			a.logger.Info().Msgf("👋 HTTP server listening on :%s", a.config.Port)
			return a.httpServer.ListenAndServe()
		}, func(err error) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := a.httpServer.Shutdown(ctx); err != nil {
				a.logger.Error().Err(err).Msg("failed to shutdown HTTP server")
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
