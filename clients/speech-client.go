package clients

import (
	"context"
	"log"

	speech "cloud.google.com/go/speech/apiv1"
)

func NewSpeechClient(ctx context.Context) *speech.Client {
	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Speech-to-Text client: %v", err)
	}
	return client
}
