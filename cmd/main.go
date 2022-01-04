package main

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

func main() {
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

	repo, err := repository.NewRepository(context.Background(), cfg.GCPProject)
	if err != nil {
		logger.Fatal("failed to initialize repository", zap.Error(err))
	}

	err = repo.ClearCollection(context.Background())
	if err != nil {
		logger.Fatal("failed to clear collection", zap.Error(err))
	}

	svc := core.NewService(&core.Config{
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

	if err = svc.FilterAndAddBalls(context.Background()); err != nil {
		logger.Fatal("failed to filter and add balls", zap.Error(err))
	}

	defer func() {
		alerterErr := a.Close()
		if alerterErr != nil {
			logger.Warn("error closing alerter", zap.Error(alerterErr))
		}
		usbcClient.Close()
	}()
}

// import (
// 	"context"
// 	"log"
// 	"os"

// 	"cloud.google.com/go/firestore"
// 	"github.com/actatum/approved-ball-list/db"
// 	"github.com/actatum/approved-ball-list/discord"
// 	"github.com/actatum/approved-ball-list/models"
// 	"github.com/actatum/approved-ball-list/usbc"
// )

// var client *firestore.Client

// func init() {
// 	var err error
// 	client, err = firestore.NewClient(context.Background(), os.Getenv("GCP_PROJECT"))
// 	if err != nil {
// 		log.Fatalf("firestore.NewClient: %v", err)
// 	}
// }

// // ApprovedBallList is the entry point for the cloud function
// // and handles orchestrating the smaller pieces to complete the workflow
// func ApprovedBallList(ctx context.Context, _ interface{}) error {
// 	// Get balls from db
// 	ballsFromDB, err := db.GetAllBalls(ctx, client)
// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	}
// 	log.Printf("Number of balls in database: %d\n", len(ballsFromDB))

// 	// Get balls from usbc
// 	ballsFromUSBC, err := usbc.GetBalls(ctx)
// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	}
// 	log.Printf("Number of approved balls from USBC: %d\n", len(ballsFromUSBC))

// 	// filter out balls from usbc that are in db
// 	result := filter(ballsFromDB, ballsFromUSBC)
// 	log.Printf("Number of newly approved balls: %d\n", len(result))

// 	if len(result) == 0 {
// 		return nil
// 	}

// 	err = discord.SendNewBalls(result)
// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	}

// 	// Add new balls to db
// 	err = db.AddBalls(ctx, client, result)
// 	if err != nil {
// 		log.Println(err)
// 		return err
// 	}

// 	return nil
// }

// func filter(fromDB []models.Ball, fromUSBC []models.Ball) []models.Ball {
// 	var unique []models.Ball

// 	for _, b := range fromUSBC { // each ball from the usbc
// 		found := false
// 		for i := 0; i < len(fromDB); i++ {
// 			if b.Name == fromDB[i].Name && b.Brand == fromDB[i].Brand {
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			unique = append(unique, b)
// 		}
// 	}

// 	return unique
// }
