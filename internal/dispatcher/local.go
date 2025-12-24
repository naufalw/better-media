package dispatcher

import (
	"better-media/internal/storage"
	"better-media/internal/transcoder"
	"better-media/pkg/models"
	"context"
	"log"

	"github.com/google/uuid"
)

type LocalDispatcher struct {
	s3Client *storage.S3Client
}

func NewLocalDispatcher(s3Client *storage.S3Client) *LocalDispatcher {
	return &LocalDispatcher{
		s3Client,
	}
}

// Dispatch directly to the FFMpeg in this process
func (d *LocalDispatcher) Dispatch(ctx context.Context, payload models.VideoEncodingPayload) (string, error) {
	jobID := uuid.New().String()

	go func() {
		pipeline, err := transcoder.NewEncodingPipeline(payload)

		if err != nil {
			log.Printf("[%s] Failed to create pipeline :%v", jobID, err)
			return
		}

		if err := pipeline.Run(context.Background(), d.s3Client); err != nil {
			log.Printf("[%s] Pipeline failed: %v", jobID, err)
			return
		}

		log.Printf("[%s] Pipeline completed successfully", jobID)

	}()

	return jobID, nil
}
