package repository

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/actatum/approved-ball-list/core"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

const ballCollection = "balls"
const batchSize = 500

// Repository handles interfacing with the persistence in layer
// in this case implemented by firestore
type Repository struct {
	client *firestore.Client
}

// NewRepository creates a new repository connected to the given GCP project via the project ID
func NewRepository(ctx context.Context, projectID string) (*Repository, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("firestore.NewClient: %w", err)
	}

	return &Repository{
		client: client,
	}, nil
}

// GetAllBalls retreives all the balls from the firestore database
func (r *Repository) GetAllBalls(ctx context.Context) ([]core.Ball, error) {
	list := make([]core.Ball, 0)
	iter := r.client.Collection(ballCollection).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iter.Next: %w", err)
		}
		var b core.Ball
		if err = doc.DataTo(&b); err != nil {
			return nil, fmt.Errorf("doc.DataTo: %w", err)
		}
		list = append(list, b)
	}

	return list, nil
}

// InsertNewBalls adds the new balls to the database
func (r *Repository) InsertNewBalls(ctx context.Context, balls []core.Ball) error {
	if len(balls) == 0 {
		return nil
	}

	batch := r.client.Batch()
	currentBatch := 0
	for i := 0; i < len(balls); i++ {
		currentBatch++
		ref := r.client.Collection(ballCollection).Doc(uuid.NewString())
		batch.Set(ref, balls[i])
		if currentBatch == batchSize {
			_, err := batch.Commit(ctx)
			if err != nil {
				return fmt.Errorf("batch.Commit: %w", err)
			}
			currentBatch = 0
			batch = r.client.Batch()
		}
	}
	_, err := batch.Commit(ctx)
	if err != nil {
		return fmt.Errorf("batch.Commit: %w", err)
	}

	return nil
}

// ClearCollection drops the entire collection of bowling balls
func (r *Repository) ClearCollection(ctx context.Context) error {
	ref := r.client.Collection(ballCollection)
	for {
		// Get a batch of documents
		iter := ref.Limit(batchSize).Documents(ctx)
		numDeleted := 0

		// Iterate through the documents, adding
		// a delete operation for each one to a
		// WriteBatch.
		batch := r.client.Batch()
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

// Close shuts down the underlying firestore client
func (r *Repository) Close() error {
	return r.client.Close()
}
