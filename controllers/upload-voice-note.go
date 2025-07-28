package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/zerpajose/resume-my-mom-voice-notes-service-go/helpers"
	"github.com/zerpajose/resume-my-mom-voice-notes-service-go/types"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

func UploadVoiceNote(
	ctx context.Context,
	firestoreClient *firestore.Client,
	storageClient *storage.Client,
	bucketName string,
	collectionName string,
	chatId string,
	fileHeader *multipart.FileHeader,
) error {
	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Read file into buffer
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Generate file name
	ext := ""
	if parts := types.GetFileExtension(fileHeader.Filename); parts != "" {
		ext = parts
	}
	fileName := fmt.Sprintf("%s/%s.%s", chatId, uuid.New().String(), ext)

	// Upload to storage
	if err := helpers.UploadStorageFile(ctx, storageClient, bucketName, fileName, &buf); err != nil {
		return fmt.Errorf("failed to upload file to storage: %w", err)
	}

	// Firestore logic
	collection := firestoreClient.Collection(collectionName)
	docs, err := collection.
		Where("chatId", "==", chatId).
		Where("finished", "==", false).
		Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to query firestore: %w", err)
	}

	if len(docs) == 0 {
		id := uuid.New().String()
		_, err := collection.Doc(id).Set(ctx, map[string]interface{}{
			"id":        id,
			"fileKeys":  []string{fileName},
			"chatId":    chatId,
			"createdAt": time.Now().Format(time.RFC3339),
			"finished":  false,
		})
		return err
	}

	// Update existing doc
	doc := docs[0]
	data := doc.Data()
	fileKeys := []string{}
	if fk, ok := data["fileKeys"].([]interface{}); ok {
		for _, v := range fk {
			if s, ok := v.(string); ok {
				fileKeys = append(fileKeys, s)
			}
		}
	}
	fileKeys = append(fileKeys, fileName)
	_, err = doc.Ref.Update(ctx, []firestore.Update{
		{Path: "fileKeys", Value: fileKeys},
		{Path: "updatedAt", Value: time.Now().Format(time.RFC3339)},
	})
	return err
}
