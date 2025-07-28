package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/zerpajose/resume-my-mom-voice-notes-service-go/clients"
	"github.com/zerpajose/resume-my-mom-voice-notes-service-go/config"
)

// GeminiRequest is the request payload for the Gemini API.
type GeminiRequest struct {
	Model    string `json:"model"`
	Contents string `json:"contents"`
}

// GeminiResponse is the response from the Gemini API.
type GeminiResponse struct {
	Text string `json:"text"`
}

// AIResume summarizes the transcription using Gemini API.
func AIResume(ctx context.Context, cfg config.Config, transcription string) (string, error) {
	// Get Gemini API key from Secret Manager
	apiKey, err := clients.GetSecret(ctx, cfg.GoogleProjectNumber, cfg.GeminiAPIKey)
	if err != nil {
		return "", fmt.Errorf("failed to get Gemini API key: %w", err)
	}

	// Prepare prompt
	prompt := fmt.Sprintf(`Quiero que resumas la siguiente transcripción de voz en una
lista de los puntos abordados mas importantes y separarlo en secciones si
se toca mas de un tema, aquí está la transcripción: "%s".`, transcription)

	// Prepare request
	reqBody, err := json.Marshal(GeminiRequest{
		Model:    "gemini-2.5-flash",
		Contents: prompt,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request to Gemini API
	req, err := http.NewRequestWithContext(ctx, "POST", "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini API error: %s", string(body))
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("failed to decode Gemini response: %w", err)
	}

	return geminiResp.Text, nil
}
