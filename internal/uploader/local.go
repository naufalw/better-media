package uploader

import (
	"better-media/internal/database"
	"better-media/internal/storage"
	"context"
	"log"
)

type LocalUploader struct {
	s3Client *storage.S3Client
	db       *database.DB
	jobID    string
}

func NewLocalUploader(s3Client *storage.S3Client, db *database.DB, jobID string) *LocalUploader {
	return &LocalUploader{
		s3Client: s3Client,
		db:       db,
		jobID:    jobID,
	}
}

// Upload the file by simply calling the s3 client
func (u *LocalUploader) Upload(ctx context.Context, localPath, objectKey string) error {
	return u.s3Client.UploadFile(ctx, localPath, objectKey)
}

func (u *LocalUploader) NotifyPlayable() error {
	log.Printf("[%s] First rendition ready - video is now playable", u.jobID)
	return u.db.UpdateJobStatus(u.jobID, "playable", nil)
}

func (u *LocalUploader) NotifyComplete() error {
	log.Printf("[%s] Pipeline completed successfully", u.jobID)
	return u.db.UpdateJobStatus(u.jobID, "completed", nil)
}
func (u *LocalUploader) NotifyFailed(errMsg string) error {
	log.Printf("[%s] Pipeline failed: %s", u.jobID, errMsg)
	return u.db.UpdateJobStatus(u.jobID, "failed", &errMsg)
}

func (u *LocalUploader) UpdateProgress(percent int) error {
	return u.db.UpdateJobProgress(u.jobID, percent)
}
