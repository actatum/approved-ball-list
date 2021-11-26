package p

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/actatum/approved-ball-list/db"
	"github.com/actatum/approved-ball-list/discord"
	"github.com/actatum/approved-ball-list/models"
	"github.com/actatum/approved-ball-list/usbc"
)

var client *firestore.Client

func init() {
	var err error
	client, err = firestore.NewClient(context.Background(), os.Getenv("GCP_PROJECT"))
	if err != nil {
		log.Fatalf("firestore.NewClient: %v", err)
	}
}

// ApprovedBallList is the entry point for the cloud function
// and handles orchestrating the smaller pieces to complete the workflow
func ApprovedBallList(ctx context.Context, _ interface{}) error {
	// Get balls from db
	ballsFromDB, err := db.GetAllBalls(ctx, client)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Number of balls in database: %d\n", len(ballsFromDB))

	// Get balls from usbc
	ballsFromUSBC, err := usbc.GetBalls(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Number of approved balls from USBC: %d\n", len(ballsFromUSBC))

	// filter out balls from usbc that are in db
	result := filter(ballsFromDB, ballsFromUSBC)
	log.Printf("Number of newly approved balls: %d\n", len(result))

	err = discord.SendNewBalls(result)
	if err != nil {
		log.Println(err)
		return err
	}

	// Add new balls to db
	err = db.AddBalls(ctx, client, result)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func filter(fromDB []models.Ball, fromUSBC []models.Ball) []models.Ball {
	var unique []models.Ball

	for _, b := range fromUSBC { // each ball from the usbc
		found := false
		for i := 0; i < len(fromDB); i++ {
			if b.Name == fromDB[i].Name {
				found = true
				break
			}
		}
		if !found {
			unique = append(unique, b)
		}
	}

	return unique
}
