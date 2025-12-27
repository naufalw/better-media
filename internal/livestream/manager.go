package livestream

import (
	"better-media/internal/database"
	"better-media/internal/dispatcher"
	"better-media/internal/storage"
	"better-media/pkg/models"
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type Manager struct {
	RTMPServer *RTMPServer
	S3Client   *storage.S3Client
	DB         *database.DB
	Dispatcher dispatcher.JobDispatcher
}

// Create new livestream manager
func NewManager(rtmpPort int, outputBase string, s3Client *storage.S3Client, db *database.DB, disp dispatcher.JobDispatcher) *Manager {
	m := &Manager{
		S3Client:   s3Client,
		DB:         db,
		Dispatcher: disp,
	}
	rtmpServer := NewRTMPServer(rtmpPort, outputBase)

	rtmpServer.OnStreamStart = m.onStreamStart
	rtmpServer.OnStreamEnd = m.onStreamEnd
	rtmpServer.ValidateKey = m.validateKey
	rtmpServer.OnRecordReady = m.onRecordReady
	m.RTMPServer = rtmpServer
	return m
}

func (m *Manager) Start() error {
	return m.RTMPServer.Start()
}
func (m *Manager) onStreamStart(streamKey, hlsPath string) {
	log.Printf("[LivestreamManager] Stream started: %s -> %s", streamKey, hlsPath)
}
func (m *Manager) onStreamEnd(streamKey string) {
	log.Printf("[LivestreamManager] Stream ended: %s", streamKey)
}
func (m *Manager) validateKey(key string) bool {
	streamKey, err := m.DB.ValidateStreamKey(key)
	return err == nil && streamKey != nil
}
func (m *Manager) onRecordReady(streamKey, outputDir string) {
	go m.uploadRecording(streamKey, outputDir)
}
func (m *Manager) uploadRecording(streamKey, outputDir string) {
	log.Printf("[LivestreamManager] Processing recording: %s", streamKey)
	recordingPath := filepath.Join(outputDir, "recording.mkv")
	if _, err := os.Stat(recordingPath); os.IsNotExist(err) {
		log.Printf("[LivestreamManager] No recording found at %s", recordingPath)
		return
	}
	videoID := uuid.New().String()
	objectKey := filepath.Join(videoID, "source", "recording.mkv")

	// Upload to S3
	if err := m.S3Client.UploadFile(context.Background(), recordingPath, objectKey); err != nil {
		log.Printf("[LivestreamManager] Upload failed: %v", err)
		return
	}

	log.Printf("[LivestreamManager] Recording uploaded: %s", objectKey)
	log.Printf("[LivestreamManager] Video ID: %s", videoID)

	// TODO: LIBRARY ID
	m.DB.CreateVideo(videoID, "Livestream Recording", "livestream", nil)

	// transcode
	if m.Dispatcher != nil {
		payload := models.VideoEncodingPayload{
			VideoID:     videoID,
			InputFile:   "recording.mkv",
			Resolutions: []int{360, 720},
		}
		jobID, err := m.Dispatcher.Dispatch(context.Background(), payload)
		if err != nil {
			log.Printf("[LivestreamManager] Failed to dispatch transcoding: %v", err)
		} else {
			log.Printf("[LivestreamManager] Transcoding job started: %s", jobID)
		}
	}

	os.RemoveAll(outputDir)
}

// Return active stream keys
func (m *Manager) GetActiveStreams() []string {
	return m.RTMPServer.GetActiveStreams()
}

// Return  stream info
func (m *Manager) GetStream(key string) *Stream {
	return m.RTMPServer.GetStream(key)
}
