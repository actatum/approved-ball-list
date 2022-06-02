package main

import (
	"context"
	"fmt"
	"net/http"

	abl2 "github.com/actatum/approved-ball-list/abl"
	"github.com/actatum/approved-ball-list/cockroachdb"
	"github.com/actatum/approved-ball-list/discord"
	"github.com/actatum/approved-ball-list/usbc"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

// NOTE: This main is only for local testing.

func main() {
	logger := abl2.NewLogger("test", zerolog.DebugLevel)

	db, err := cockroachdb.NewDB("postgres://root@localhost:26257/defaultdb")
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	defer func() {
		_ = db.Close()
	}()

	dg, err := discordgo.New("Bot OTExMDgwODE0ODgyNzI1OTM5.YZcMIQ.27D08TjxbbFs2EFBTW-a9w0NqQM")
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	defer func() {
		_ = dg.Close()
	}()

	ballService := cockroachdb.NewBallService(db)
	notiService := discord.NewNotificationService(dg, []string{"77164508690264064"})

	client := usbc.NewClient(&usbc.Config{
		Logger:     logger,
		HTTPClient: &http.Client{},
	})

	usbcList, err := client.GetApprovedBallList(context.Background())
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	list, cnt, err := ballService.ListBalls(context.Background(), abl2.BallFilter{})
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	if len(usbcList) > cnt {
		fmt.Println("NEWLY APPROVED BALL(S)")
		newlyApproved := make([]abl2.Ball, 0)
		for _, b := range usbcList {
			if !contains(list, b) {
				logger.Info().Msgf("new ball: %s %s", b.Brand, b.Name)
				newlyApproved = append(newlyApproved, b)
			}
		}

		if err = ballService.AddBalls(context.Background(), newlyApproved...); err != nil {
			logger.Fatal().Err(err).Send()
		}

		// send notifications
		notifications := make([]abl2.Notification, 0, len(newlyApproved))
		for _, b := range newlyApproved {
			notifications = append(notifications, abl2.Notification{Ball: b})
		}

		if err = notiService.SendNotifications(context.Background(), notifications...); err != nil {
			logger.Fatal().Err(err).Send()
		}

	} else if len(usbcList) < cnt {
		revoked := make([]abl2.Ball, 0)
		for _, b := range list {
			if !contains(usbcList, b) {
				logger.Info().Msgf("ball revoked: %s %s", b.Brand, b.Name)
				revoked = append(revoked, b)
			}
		}

		if err = ballService.RemoveBalls(context.Background(), revoked...); err != nil {
			logger.Fatal().Err(err).Send()
		}
	}
}

func contains(s []abl2.Ball, b abl2.Ball) bool {
	for _, a := range s {
		if a.Equal(b) {
			return true
		}
	}
	return false
}
