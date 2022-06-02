package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/actatum/approved-ball-list/abl"
	"github.com/actatum/approved-ball-list/cockroachdb"
	"github.com/actatum/approved-ball-list/discord"
	"github.com/actatum/approved-ball-list/usbc"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

// NOTE: This main is only for local testing.

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	logger := abl.NewLogger("test", zerolog.DebugLevel)

	db, err := cockroachdb.NewDB("postgres://root@localhost:26257/defaultdb")
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	defer func() {
		_ = db.Close()
	}()

	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	defer func() {
		_ = dg.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ballService := cockroachdb.NewBallService(db)
	notiService := discord.NewNotificationService(dg, []string{os.Getenv("PERSONAL_CHANNEL_ID")})

	usbcClient := usbc.NewClient(&usbc.Config{
		Logger:     logger,
		HTTPClient: &http.Client{},
	})

	usbcList, err := usbcClient.GetApprovedBallList(ctx)
	if err != nil {
		return err
	}

	// usbcList = usbcList[2:] uncomment to test revoking from usbclist

	list, _, err := ballService.ListBalls(ctx, abl.BallFilter{})
	if err != nil {
		return err
	}

	newlyApproved := make([]abl.Ball, 0)
	for _, b := range usbcList {
		if !contains(list, b) {
			logger.Info().Msgf("new ball: %s %s", b.Brand, b.Name)
			newlyApproved = append(newlyApproved, b)
		}
	}
	logger.Info().Msgf("%d newly approved balls", len(newlyApproved))

	if err = ballService.AddBalls(ctx, newlyApproved...); err != nil {
		return err
	}

	revoked := make([]abl.Ball, 0)
	for _, b := range list {
		if !contains(usbcList, b) {
			logger.Info().Msgf("ball revoked: %s %s", b.Brand, b.Name)
			revoked = append(revoked, b)
		}
	}
	logger.Info().Msgf("%d balls revoked", len(revoked))

	if err = ballService.RemoveBalls(ctx, revoked...); err != nil {
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
