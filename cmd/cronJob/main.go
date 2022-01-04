package p

import (
	"context"
	"fmt"
	"os"

	"github.com/actatum/approved-ball-list/alerter"
	"github.com/actatum/approved-ball-list/config"
	"github.com/actatum/approved-ball-list/core"
	"github.com/actatum/approved-ball-list/log"
	"github.com/actatum/approved-ball-list/repository"
	"github.com/actatum/approved-ball-list/usbc"
	"go.uber.org/zap"
)

var svc core.Service

func init() {
	logger, err := log.NewLogger()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg, err := config.NewAppConfig()
	if err != nil {
		logger.Fatal("failed to initialize app config", zap.Error(err))
	}

	a, err := alerter.NewAlerter(cfg.DiscordToken)
	if err != nil {
		logger.Fatal("failed to initialize alerter", zap.Error(err))
	}

	usbcClient := usbc.NewClient(&usbc.Config{
		Logger:     logger,
		HTTPClient: nil,
	})

	repo, err := repository.NewRepository(context.Background(), cfg.GCPProjectID)
	if err != nil {
		logger.Fatal("failed to initialize repository", zap.Error(err))
	}

	svc = core.NewService(&core.Config{
		Logger: logger,
		DiscordChannels: map[string]core.DiscordChannel{
			"motivated": {
				Name:   "motivated",
				ID:     cfg.MotivatedChannelID,
				Brands: []string{"Motiv"},
			},
			"panda-pack": {
				Name:   "panda-pack",
				ID:     cfg.PandapackChannelID,
				Brands: []string{"Storm", "Roto Grip", "900 Global"},
			},
			"brunswick-central": {
				Name:   "brunswick-central",
				ID:     cfg.BrunswickCentralChannelID,
				Brands: []string{"Brunswick", "Columbia", "DV8", "Ebonite", "Hammer", "Radical", "Track"},
			},
			"personal": {
				Name:   "personal channel",
				ID:     cfg.PersonalChannelID,
				Brands: []string{"900 Global", "BIG Bowling", "Brunswick", "Columbia", "DV8", "Ebonite", "Hammer", "Motiv", "Radical", "Roto Grip", "Storm", "Track Inc."},
			},
		},
		Repository: repo,
		Alerter:    a,
		USBC:       usbcClient,
	})
}

// CronJob is the entry point for the cronjob cloud function.
// This function retrieves balls from the usbc and our database compares them to find
// new entries on the usbc approved ball list and writes the new entries to our database
func CronJob(ctx context.Context, _ interface{}) error {
	return svc.FilterAndAddBalls(ctx)
}
