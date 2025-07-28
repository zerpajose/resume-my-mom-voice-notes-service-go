package clients

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
)

func NewFirestoreClient(ctx context.Context, projectID string) *firestore.Client {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	return client
}
