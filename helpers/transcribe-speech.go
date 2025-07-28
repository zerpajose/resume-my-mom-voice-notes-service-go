package helpers

import (
	"context"
	"fmt"

	speech "cloud.google.com/go/speech/apiv1"
	speechpb "cloud.google.com/go/speech/apiv1/speechpb"
	"cloud.google.com/go/storage"
)

func TranscribeSpeech(
	ctx context.Context,
	speechClient *speech.Client,
	storageClient *storage.Client,
	bucketName string,
	fileKey string,
) (string, error) {
	// Build the request
	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Model:                 "latest_long",
			Encoding:              speechpb.RecognitionConfig_MP3,
			SampleRateHertz:       48000,
			AudioChannelCount:     1,
			EnableWordTimeOffsets: true,
			EnableWordConfidence:  true,
			LanguageCode:          "es-US",
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Uri{
				Uri: fmt.Sprintf("gs://%s/%s", bucketName, fileKey),
			},
		},
	}

	op, err := speechClient.LongRunningRecognize(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to start recognition: %w", err)
	}

	resp, err := op.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("recognition failed: %w", err)
	}

	if len(resp.Results) == 0 {
		return "", fmt.Errorf("no transcription results found")
	}

	var transcription string
	for _, result := range resp.Results {
		if len(result.Alternatives) == 0 {
			return "", fmt.Errorf("no alternatives found for result")
		}
		transcription += result.Alternatives[0].Transcript + "\n"
	}

	// Delete the file from GCS after transcription
	_ = storageClient.Bucket(bucketName).Object(fileKey).Delete(ctx)
	// Ignore error if file is already deleted

	return transcription, nil
}
