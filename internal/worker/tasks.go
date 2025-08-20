package worker

import (
	"better-media/internal/storage"
	"better-media/pkg/models"
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
)

// The primary motivation for this is to simplify the dependency injection during task processing
// This allows us to pass the S3 client from the parent function, and we dont need to destructure the handler
// Refer to how we pass s3 on the main function in cmd/worker/main.go
type TaskProcessor struct {
	S3Client *storage.S3Client
}

func NewTaskProcessor(s3c *storage.S3Client) *TaskProcessor {
	return &TaskProcessor{S3Client: s3c}
}

func (processor *TaskProcessor) HandleVideoEncodeTask(ctx context.Context, t *asynq.Task) error {
	var payload models.VideoEncodingPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	log.Printf("Starting pipeline for VideoID: %s", payload.VideoID)

	pipeline, err := NewEncodingPipeline(payload)
	if err != nil {
		log.Printf("!!! PIPELINE FAILED for VideoID %s: could not create pipeline: %v", payload.VideoID, err)
		return err
	}

	if err := pipeline.Run(ctx, processor.S3Client); err != nil {
		log.Printf("!!! PIPELINE FAILED for VideoID %s: %v", payload.VideoID, err)
		return err
	}

	return nil
}
