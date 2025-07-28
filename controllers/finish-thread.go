package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/zerpajose/resume-my-mom-voice-notes-service-go/clients"
	"github.com/zerpajose/resume-my-mom-voice-notes-service-go/helpers"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
)

func FinishThread(
	ctx context.Context,
	firestoreClient *firestore.Client,
	storageClient *storage.Client,
	bucketName string,
	collectionName string,
	chatId string,
	aiResumeFunc func(string) (string, error), // Pass your AI summary function here
) (string, error) {
	collection := firestoreClient.Collection(collectionName)
	iter := collection.
		Where("chatId", "==", chatId).
		Where("finished", "==", false).
		Documents(ctx)
	docs, err := iter.GetAll()
	if err != nil {
		return "", fmt.Errorf("failed to query Firestore: %w", err)
	}
	if len(docs) == 0 {
		return "", fmt.Errorf("no open thread found")
	}
	if len(docs) > 1 {
		return "", fmt.Errorf("multiple open threads found, please resolve this issue")
	}

	doc := docs[0]
	docData := doc.Data()
	fileKeysIface, ok := docData["fileKeys"].([]interface{})
	if !ok || len(fileKeysIface) == 0 {
		return "", fmt.Errorf("no files to merge in this thread")
	}
	// Convert []interface{} to []string
	fileKeys := make([]string, 0, len(fileKeysIface))
	for _, v := range fileKeysIface {
		if s, ok := v.(string); ok {
			fileKeys = append(fileKeys, s)
		}
	}

	// Merge audio files
	mergedFileKey, err := helpers.MergeAudioFiles(ctx, storageClient, bucketName, fileKeys, "mp3")
	if err != nil {
		return "", fmt.Errorf("failed to merge audio files: %w", err)
	}

	// Transcribe merged audio
	transcription, err := helpers.TranscribeSpeech(ctx, clients.NewSpeechClient(ctx), storageClient, bucketName, mergedFileKey)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe speech: %w", err)
	}

	// Summarize transcription
	summary, err := aiResumeFunc(transcription)
	if err != nil {
		return "", fmt.Errorf("failed to summarize transcription: %w", err)
	}

	// Mark thread as finished
	_, err = doc.Ref.Update(ctx, []firestore.Update{
		{Path: "finished", Value: true},
		{Path: "finishedAt", Value: time.Now().Format(time.RFC3339)},
	})
	if err != nil {
		return "", fmt.Errorf("failed to update Firestore: %w", err)
	}

	return summary, nil
}
