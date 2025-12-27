package uploader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type RemoteUploader struct {
	callbackURL string
	jobID       string
	httpClient  *http.Client
}

func NewRemoteUploader(callbackURL, jobID string) *RemoteUploader {
	return &RemoteUploader{
		callbackURL: callbackURL,
		jobID:       jobID,
		httpClient:  &http.Client{},
	}
}

// Upload the video to s3, but request the presign from main api
func (u *RemoteUploader) Upload(ctx context.Context, localPath, objectKey string) error {

	// we get the PUT url from main api
	presignURL := fmt.Sprintf("%s/v1/callbacks/%s/presign-upload", u.callbackURL, u.jobID)

	reqBody, _ := json.Marshal(map[string][]string{
		"files": {objectKey},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", presignURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create presign request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request presigned URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("presign request failed with status %d", resp.StatusCode)
	}

	var presignResp struct {
		URLs map[string]string `json:"urls"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&presignResp); err != nil {
		return fmt.Errorf("failed to decode presign response: %w", err)
	}

	uploadURL, ok := presignResp.URLs[objectKey]
	if !ok {
		return fmt.Errorf("no presigned URL returned for %s", objectKey)
	}

	// ====== here we have uploadURL, just upload

	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	uploadReq, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, file)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	uploadReq.ContentLength = fileInfo.Size()

	uploadResp, err := u.httpClient.Do(uploadReq)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode >= 400 {
		body, _ := io.ReadAll(uploadResp.Body)
		return fmt.Errorf("upload failed with status %d: %s", uploadResp.StatusCode, string(body))
	}

	log.Printf("[%s] Uploaded %s", u.jobID, objectKey)
	return nil

}

// Update the status of current video
func (u *RemoteUploader) updateStatus(status string, errMsg *string) error {
	statusURL := fmt.Sprintf("%s/v1/callbacks/%s/status", u.callbackURL, u.jobID)

	body := map[string]interface{}{"status": status}
	if errMsg != nil {
		body["error"] = *errMsg
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal status update: %w", err)
	}

	req, err := http.NewRequest("POST", statusURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create status request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("status update failed with status %d", resp.StatusCode)
	}

	return nil
}

func (u *RemoteUploader) NotifyPlayable() error {
	log.Printf("[%s] Notifying API: playable", u.jobID)
	return u.updateStatus("playable", nil)
}
func (u *RemoteUploader) NotifyComplete() error {
	log.Printf("[%s] Notifying API: completed", u.jobID)
	return u.updateStatus("completed", nil)
}
func (u *RemoteUploader) NotifyFailed(errMsg string) error {
	log.Printf("[%s] Notifying API: failed - %s", u.jobID, errMsg)
	return u.updateStatus("failed", &errMsg)
}

func (u *RemoteUploader) UpdateProgress(percent int) error {
	statusURL := fmt.Sprintf("%s/v1/callbacks/%s/progress", u.callbackURL, u.jobID)

	body := map[string]int{"progress": percent}
	reqBody, _ := json.Marshal(body)

	resp, err := u.httpClient.Post(statusURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("progress update failed with status %d", resp.StatusCode)
	}

	return nil
}

func (u *RemoteUploader) UpdateMetadata(videoID string, durationMs int64, fileSizeBytes int64, width, height int) error {
	url := fmt.Sprintf("%s/v1/callbacks/%s/metadata", u.callbackURL, u.jobID)

	body := map[string]interface{}{
		"video_id":          videoID,
		"duration_ms":       durationMs,
		"file_size_bytes":   fileSizeBytes,
		"resolution_width":  width,
		"resolution_height": height,
	}

	reqBody, _ := json.Marshal(body)
	resp, err := u.httpClient.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("metadata update failed with status %d", resp.StatusCode)
	}
	return nil
}
