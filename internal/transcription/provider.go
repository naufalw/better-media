package transcription

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Segment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

type TranscriptionResult struct {
	Language string    `json:"language"`
	Text     string    `json:"text"`
	Segments []Segment `json:"segments"`
}

type Provider interface {
	Transcribe(ctx context.Context, audioPath string) (*TranscriptionResult, error)
}

type OpenAICompatibleProvider struct {
	BaseURL    string // e.g., "https://api.groq.com/openai/v1" or "https://api.openai.com/v1"
	APIKey     string
	Model      string // e.g., "whisper-large-v3" or "whisper-1"
	HTTPClient *http.Client
}

func NewProvider(baseURL, apiKey, model string) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

func (p *OpenAICompatibleProvider) Transcribe(ctx context.Context, audioPath string) (*TranscriptionResult, error) {
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// request struct
	writer.WriteField("model", p.Model)
	writer.WriteField("response_format", "verbose_json")
	writer.WriteField("timestamp_granularities[]", "segment")
	writer.Close()

	url := fmt.Sprintf("%s/audio/transcriptions", p.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("transcription request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("transcription API error %d: %s", resp.StatusCode, string(body))
	}

	var result TranscriptionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil

}
