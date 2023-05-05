package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/actatum/approved-ball-list/internal/abl"
	"github.com/actatum/approved-ball-list/internal/crdb"
	"github.com/actatum/approved-ball-list/internal/discord"
	"github.com/actatum/approved-ball-list/internal/gcs"
	"github.com/actatum/approved-ball-list/internal/mocks"
	"github.com/actatum/approved-ball-list/internal/sqlite"
	"github.com/actatum/approved-ball-list/internal/usbc"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"

	// imported for side effects
	_ "github.com/jackc/pgx/v5/stdlib"
)

// CLI runs the command line interface.
func CLI(args []string) int {
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
				fmt.Printf("NewBackupManager: %v", err)
				return 1
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

	var db *sqlx.DB
	{
		db, err = sqlx.Connect("pgx", cfg.CockroachDBURL)
		if err != nil {
			fmt.Printf("sqlx.Connect: %v", err)
			return 1
		}
		defer db.Close()
	}

	var repo abl.Repository
	{
		repo, err = crdb.NewRepository(db)
		if err != nil {
			fmt.Printf("NewRepository: %v", err)
			return 1
		}
	}

	var notifier abl.Notifier
	{
		if cfg.Env != "local" {
			notifier, err = discord.NewNotifier(cfg.DiscordToken, cfg.DiscordChannels)
			if err != nil {
				fmt.Printf("NewNotifier: %v", err)
				return 1
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
	fmt.Println(svc)

	app := &cli.App{
		Name:  "abl",
		Usage: "interact with the USBC approved ball list",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Value:   "",
				Usage:   "output format",
				Aliases: []string{"o"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "list approved balls",
				Action: func(c *cli.Context) error {
					// svc.RefreshBalls(c.Context)
					if c.String("output") == "json" {
						fmt.Println("output json")
					}

					result, err := repo.ListBalls(c.Context, abl.BallFilter{
						PageSize: 15,
					})
					if err != nil {
						return err
					}

					fmt.Println(result)

					if c.String("output") == "json" {
						data, err := json.MarshalIndent(result, "", "\t")
						if err != nil {
							return err
						}

						fmt.Println(string(data))
					} else {
						fmt.Println(result)
					}

					return nil
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	if err := app.Run(args); err != nil {
		return 1
	}

	return 0
}
