package clients

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
)

func NewStorageClient(ctx context.Context) *storage.Client {
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Storage client: %v", err)
	}
	return client
}
