package dispatcher

import (
	"better-media/internal/storage"
	"better-media/pkg/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type HTTPDispatcher struct {
	workerUrl   string
	httpClient  *http.Client
	s3Client    *storage.S3Client
	callbackUrl string
}

func NewHTTPDispatcher(s3Client *storage.S3Client, workerURL, callbackURL string) *HTTPDispatcher {

	return &HTTPDispatcher{
		workerUrl:   workerURL,
		httpClient:  &http.Client{},
		s3Client:    s3Client,
		callbackUrl: callbackURL,
	}
}

func (d *HTTPDispatcher) Dispatch(ctx context.Context, payload models.VideoEncodingPayload) (string, error) {
	jobID := uuid.New().String()

	validDuration := time.Minute * 15

	presignedGet, err := d.s3Client.GeneratePresignedGet(ctx, fmt.Sprintf("%s/source/%s", payload.VideoID, payload.InputFile), validDuration)

	if err != nil {
		return "", fmt.Errorf("Error in dispatching to external %v", err)
	}

	transcodeRequest := models.TranscodeRequest{JobID: jobID, VideoID: payload.VideoID, DownloadURL: presignedGet.URL, CallbackURL: d.callbackUrl, Resolutions: payload.Resolutions}

	jsonBody, err := json.Marshal(transcodeRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.workerUrl, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("worker returned error status: %d", resp.StatusCode)
	}
	return jobID, nil

}
