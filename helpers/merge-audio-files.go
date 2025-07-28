package helpers

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"cloud.google.com/go/storage"

	"github.com/google/uuid"
)

// MergeAudioFiles downloads files from GCS, merges them using ffmpeg, uploads the result, deletes originals, and returns the merged file key.
func MergeAudioFiles(ctx context.Context, storageClient *storage.Client, bucketName string, fileKeys []string, outputFormat string) (string, error) {
	tmpDir := filepath.Join(os.TempDir(), "merge-audio")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create tmp dir: %w", err)
	}

	var localFiles []string
	for _, key := range fileKeys {
		localPath := filepath.Join(tmpDir, key)
		if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
			return "", fmt.Errorf("failed to create parent dir: %w", err)
		}
		f, err := os.Create(localPath)
		if err != nil {
			return "", fmt.Errorf("failed to create local file: %w", err)
		}
		rc, err := storageClient.Bucket(bucketName).Object(key).NewReader(ctx)
		if err != nil {
			f.Close()
			return "", fmt.Errorf("failed to open GCS file: %w", err)
		}
		if _, err := io.Copy(f, rc); err != nil {
			rc.Close()
			f.Close()
			return "", fmt.Errorf("failed to download file: %w", err)
		}
		rc.Close()
		f.Close()
		localFiles = append(localFiles, localPath)
	}

	// Prepare output file
	mergedFileName := fmt.Sprintf("merged-%s.%s", uuid.New().String(), outputFormat)
	mergedFilePath := filepath.Join(tmpDir, mergedFileName)

	// Create ffmpeg input list file
	listFile := filepath.Join(tmpDir, "inputs.txt")
	list, err := os.Create(listFile)
	if err != nil {
		return "", fmt.Errorf("failed to create ffmpeg list file: %w", err)
	}
	for _, f := range localFiles {
		fmt.Fprintf(list, "file '%s'\n", f)
	}
	list.Close()

	// Merge using ffmpeg
	var cmd *exec.Cmd
	if outputFormat == "mp3" {
		// Re-encode to mp3 to avoid codec mismatch errors
		cmd = exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", listFile, "-c:a", "libmp3lame", mergedFilePath)
	} else {
		// Use stream copy for other formats (if input matches output)
		cmd = exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", listFile, "-c", "copy", mergedFilePath)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg error: %v, output: %s", err, string(out))
	}

	// Upload merged file to storage
	mergedFile, err := os.Open(mergedFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open merged file: %w", err)
	}
	defer mergedFile.Close()
	wc := storageClient.Bucket(bucketName).Object(mergedFileName).NewWriter(ctx)
	if _, err := io.Copy(wc, mergedFile); err != nil {
		wc.Close()
		return "", fmt.Errorf("failed to upload merged file: %w", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Delete original files from GCS
	for _, key := range fileKeys {
		_ = storageClient.Bucket(bucketName).Object(key).Delete(ctx)
	}

	// Cleanup local files
	for _, f := range localFiles {
		_ = os.Remove(f)
	}
	_ = os.Remove(mergedFilePath)
	_ = os.Remove(listFile)

	return mergedFileName, nil
}
