package sn

import (
	"context"
	"net/http"
	"time"

	"github.com/actatum/approved-ball-list/abl"
	"github.com/actatum/approved-ball-list/cockroachdb"
	"github.com/actatum/approved-ball-list/config"
	"github.com/actatum/approved-ball-list/discord"
	"github.com/actatum/approved-ball-list/usbc"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

var (
	logger      *zerolog.Logger
	db          *sqlx.DB
	dg          *discordgo.Session
	usbcClient  *usbc.Client
	ballService abl.BallService
	notiService abl.NotificationService
)

func init() {
	logger = abl.NewLogger("approved-ball-list", zerolog.InfoLevel)

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init config")
	}

	db, err = cockroachdb.NewDB(cfg.CockroachDSN)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to db")
	}

	dg, err = discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create discord session")
	}

	usbcClient = usbc.NewClient(&usbc.Config{
		Logger:     logger,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	ballService = cockroachdb.NewBallService(db)
	notiService = discord.NewNotificationService(dg, []string{cfg.PersonalChannelID, cfg.USBCApprovedBallListChannelID})
}

// CronJob is the entry point for the cronjob cloud function.
// This function retrieves balls from the usbc and our database compares them to find
// new entries on the usbc approved ball list and writes the new entries to our database.
func CronJob(ctx context.Context, _ interface{}) error {
	usbcList, err := usbcClient.GetApprovedBallList(ctx)
	if err != nil {
		return err
	}

	list, cnt, err := ballService.ListBalls(ctx, abl.BallFilter{})
	if err != nil {
		return err
	}

	if len(usbcList) > cnt {
		logger.Info().Msgf("%d newly approved balls", len(usbcList)-cnt)
		newlyApproved := make([]abl.Ball, 0)
		for _, b := range usbcList {
			if !contains(list, b) {
				logger.Info().Msgf("new ball: %s %s", b.Brand, b.Name)
				newlyApproved = append(newlyApproved, b)
			}
		}

		if err = ballService.AddBalls(ctx, newlyApproved...); err != nil {
			return err
		}

		// send notifications
		notifications := make([]abl.Notification, 0, len(newlyApproved))
		for _, b := range newlyApproved {
			notifications = append(notifications, abl.Notification{Ball: b})
		}

		if err = notiService.SendNotifications(ctx, notifications...); err != nil {
			return err
		}

	} else if len(usbcList) < cnt {
		revoked := make([]abl.Ball, 0)
		for _, b := range list {
			if !contains(usbcList, b) {
				logger.Info().Msgf("ball revoked: %s %s", b.Brand, b.Name)
				revoked = append(revoked, b)
			}
		}

		if err = ballService.RemoveBalls(context.Background(), revoked...); err != nil {
			return err
		}
	}

	return nil
}

func contains(s []abl.Ball, b abl.Ball) bool {
	for _, a := range s {
		if a.Equal(b) {
			return true
		}
	}
	return false
}
