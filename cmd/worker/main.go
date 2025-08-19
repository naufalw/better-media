package main

import (
	"better-media/pkg/models"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

const redisAddr = "127.0.0.1:6379"

func main() {

	asynqServer := asynq.NewServer(asynq.RedisClientOpt{Addr: redisAddr}, asynq.Config{
		Concurrency: 1,
	})

	mux := asynq.NewServeMux()

	mux.HandleFunc(models.TaskEncodeVideo, handleVideoEncodeTask)

	if err := asynqServer.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}

func handleVideoEncodeTask(ctx context.Context, t *asynq.Task) error {
	var p models.VideoEncodingPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	log.Printf("Starting encoding job for VideoID: %s", p.VideoID)

	log.Printf("Step 1: Downloading from S3... (Simulated)")
	time.Sleep(2 * time.Second)

	log.Printf("Step 2: Running FFmpeg... (Simulated)")

	time.Sleep(10 * time.Second)

	log.Printf("Step 3: Uploading transcoded files to S3... (Simulated)")
	time.Sleep(2 * time.Second)

	log.Printf("Successfully finished encoding job for VideoID: %s", p.VideoID)
	return nil
}
