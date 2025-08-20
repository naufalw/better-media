package worker

import (
	"better-media/pkg/models"
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
)

func HandleVideoEncodeTask(ctx context.Context, t *asynq.Task) error {
	var payload models.VideoEncodingPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	log.Printf("Starting pipeline for VideoID: %s", payload.VideoID)

	pipeline, err := NewEncodingPipeline(payload)
	if err != nil {
		return err
	}

	return pipeline.Run()
}
