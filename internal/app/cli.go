package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/actatum/approved-ball-list/internal/balls"
	"github.com/actatum/approved-ball-list/internal/crdb"
	"github.com/actatum/approved-ball-list/internal/discord"
	"github.com/actatum/approved-ball-list/internal/mocks"
	"github.com/actatum/approved-ball-list/internal/usbc"
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

	var db *sql.DB
	{
		var err error
		db, err = sql.Open("pgx", cfg.CockroachDBURL)
		if err != nil {
			fmt.Printf("connecting to database: %v", err)
			return 1
		}
		defer db.Close()
	}

	var repo balls.Repository
	{
		var err error
		repo, err = crdb.NewRepository(db)
		if err != nil {
			fmt.Printf("NewRepository: %v", err)
			return 1
		}
	}

	var notificationService balls.NotificationService
	{
		if cfg.Env != "local" {
			discordNotificationService, err := discord.NewNotifier(cfg.DiscordToken, cfg.DiscordChannels)
			if err != nil {
				fmt.Printf("NewNotifier: %v", err)
				return 1
			}
			defer func() {
				e := discordNotificationService.Close()
				if e != nil {
					logger.Info().Err(e).Msg("Notifier.Close")
				}
			}()
		} else {
			notificationService = &mocks.NotificationServiceMock{
				SendNotificationFunc: func(ctx context.Context, approvedBalls []balls.Ball) error {
					return nil
				},
			}
		}
	}

	var usbcService balls.USBCService
	{
		usbcClient := usbc.NewClient(&usbc.Config{
			Logger:     nil,
			HTTPClient: &http.Client{},
		})
		defer usbcClient.Close()
		usbcService = usbcClient
	}

	svc := balls.NewService(repo, usbcService, notificationService)
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

					// result, err := repo.ListBalls(c.Context, abl.BallFilter{
					// 	PageSize: 15,
					// })
					// if err != nil {
					// 	return err
					// }

					// fmt.Println(result)

					// if c.String("output") == "json" {
					// 	data, err := json.MarshalIndent(result, "", "\t")
					// 	if err != nil {
					// 		return err
					// 	}

					// 	fmt.Println(string(data))
					// } else {
					// 	fmt.Println(result)
					// }

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
