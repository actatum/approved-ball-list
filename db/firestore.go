package db

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/actatum/approved-ball-list/models"
	"google.golang.org/api/iterator"
)

const ballCollection = "balls"

// GetAllBalls retreives all the balls from the firestore database
func GetAllBalls(ctx context.Context, client *firestore.Client) ([]models.Ball, error) {
	list := make([]models.Ball, 0)
	iter := client.Collection(ballCollection).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iter.Next: %w", err)
		}
		var b models.Ball
		if err = doc.DataTo(&b); err != nil {
			return nil, fmt.Errorf("doc.DataTo: %w", err)
		}
		list = append(list, b)
	}

	return list, nil
}

// AddBalls adds the new balls to the database
func AddBalls(ctx context.Context, client *firestore.Client, balls []models.Ball) error {

	for i := 0; i < len(balls); i++ {
		_, _, err := client.Collection(ballCollection).Add(ctx, balls[i])
		if err != nil {
			return fmt.Errorf("client.Collection.Add: %w", err)
		}
	}

	return nil
}

// ClearCollection drops the entire collection of bowling balls
func ClearCollection(ctx context.Context, client *firestore.Client) error {
	ref := client.Collection(ballCollection)
	batchSize := 500
	for {
		// Get a batch of documents
		iter := ref.Limit(batchSize).Documents(ctx)
		numDeleted := 0

		// Iterate through the documents, adding
		// a delete operation for each one to a
		// WriteBatch.
		batch := client.Batch()
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}

			batch.Delete(doc.Ref)
			numDeleted++
		}

		// If there are no documents to delete,
		// the process is over.
		if numDeleted == 0 {
			return nil
		}

		_, err := batch.Commit(ctx)
		if err != nil {
			return err
		}
	}
}
