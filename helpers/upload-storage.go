package helpers

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

// UploadStorageFile uploads a file to the specified bucket with the given fileName.
// contents can be any io.Reader (e.g., bytes.Buffer, os.File, etc.)
func UploadStorageFile(ctx context.Context, client *storage.Client, bucketName, fileName string, contents io.Reader) error {
	wc := client.Bucket(bucketName).Object(fileName).NewWriter(ctx)
	if _, err := io.Copy(wc, contents); err != nil {
		wc.Close()
		return fmt.Errorf("failed to write to bucket: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}
