package dispatcher

import (
	"better-media/internal/database"
	"better-media/internal/storage"
	"better-media/internal/transcoder"
	"better-media/pkg/models"
	"context"
	"log"

	"github.com/google/uuid"
)

type LocalDispatcher struct {
	s3Client *storage.S3Client
	db       *database.DB
}

// NewLocalDispatcher constructs a LocalDispatcher configured with the provided S3 client and database.
func NewLocalDispatcher(s3Client *storage.S3Client, db *database.DB) *LocalDispatcher {
	return &LocalDispatcher{
		s3Client,
		db,
	}
}

// Dispatch directly to the FFMpeg in this process
func (d *LocalDispatcher) Dispatch(ctx context.Context, payload models.VideoEncodingPayload) (string, error) {
	jobID := uuid.New().String()

	if err := d.db.CreateJob(jobID, payload.VideoID); err != nil {
		return "", err
	}
	// pending til this point

	go func() {
		if err := d.db.UpdateJobStatus(jobID, "processing", nil); err != nil {
			log.Printf("[%s] Failed to update job status to processing: %v", jobID, err)
		}

		pipeline, err := transcoder.NewEncodingPipeline(payload)

		if err != nil {
			errStr := err.Error()
			log.Printf("[%s] Failed to create pipeline :%v", jobID, err)
			if updateErr := d.db.UpdateJobStatus(jobID, "failed", &errStr); updateErr != nil {
				log.Printf("[%s] Failed to update job status to failed: %v", jobID, updateErr)
			}
			return
		}

		pipeline.OnFirstRenditionReady = func() {
			d.db.UpdateJobStatus(jobID, "playable", nil)
			log.Printf("[%s] First rendition ready - video is now playable", jobID)
		}

		if err := pipeline.Run(context.Background(), d.s3Client); err != nil {
			errStr := err.Error()
			log.Printf("[%s] Pipeline failed: %v", jobID, err)
			if updateErr := d.db.UpdateJobStatus(jobID, "failed", &errStr); updateErr != nil {
				log.Printf("[%s] Failed to update job status to failed: %v", jobID, updateErr)
			}
			return
		}

		if err := d.db.UpdateJobStatus(jobID, "completed", nil); err != nil {
			log.Printf("[%s] Failed to update job status to completed: %v", jobID, err)
		}
		log.Printf("[%s] Pipeline completed successfully", jobID)
	}()

	return jobID, nil
}
