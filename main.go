package p

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	// Imported for side effects
	// _ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"github.com/actatum/approved-ball-list/db"
	"github.com/actatum/approved-ball-list/discord"
	"github.com/actatum/approved-ball-list/models"
	"github.com/actatum/approved-ball-list/usbc"
)

var client *firestore.Client

func init() {
	var err error
	client, err = firestore.NewClient(context.Background(), "project-id")
	if err != nil {
		log.Fatalf("firestore.NewClient: %v", err)
	}
}

// func main() {
// 	start := time.Now()
// 	clear := flag.Bool("clearDB", false, "set to true to clear database")
// 	flag.Parse()
// 	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
// 	defer cancel()

// 	if *clear {
// 		err := db.ClearCollection(ctx, client)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		return
// 	}
// 	// Get balls from db
// 	ballsFromDB, err := db.GetAllBalls(ctx, client)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Printf("Number of balls in database: %d\n", len(ballsFromDB))

// 	// Get balls from usbc
// 	ballsFromUSBC, err := usbc.GetBalls(ctx)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Printf("Number of approved balls from USBC: %d\n", len(ballsFromUSBC))

// 	// filter out balls from usbc that are in db
// 	result := filter(ballsFromDB, ballsFromUSBC)
// 	log.Printf("Number of newly approved balls: %d\n", len(result))

// 	// iter := client.Collection("balls").Limit(1).Documents(ctx)
// 	// for {
// 	// 	doc, err := iter.Next()
// 	// 	if err == iterator.Done {
// 	// 		break
// 	// 	}
// 	// 	if err != nil {
// 	// 		log.Fatal(err)
// 	// 	}
// 	// 	fmt.Println(doc.Data())
// 	// 	_, err = client.Collection("balls").Doc(doc.Ref.ID).Delete(ctx)
// 	// 	if err != nil {
// 	// 		log.Fatal(err)
// 	// 	}
// 	// }

// 	// Respond/Send To Discord
// 	data, err := json.MarshalIndent(result, "", "  ")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Printf("New ball data: %v\n", string(data))

// 	err = discord.SendNewBalls(result)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Add new balls to db
// 	// TODO: Uncomment this to test adding the ball to the db as well
// 	// err = db.AddBalls(ctx, client, result)
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	end := time.Now()
// 	fmt.Println(end.Sub(start))
// }

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
