package main

import (
	"better-media/internal/storage"
	"better-media/internal/worker"
	"better-media/pkg/models"
	"log"
	"os"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

const redisAddr = "127.0.0.1:6379"

func main() {
	godotenv.Load()

	log.Println("Starting transcoder worker...")

	s3Client, err := storage.NewS3Client(
		os.Getenv("S3_BUCKET_NAME"),
		os.Getenv("S3_ENDPOINT"),
		"auto",
	)

	if err != nil {
		log.Fatalf("failed to create s3 client: %v", err)
	}

	asynqServer := asynq.NewServer(asynq.RedisClientOpt{Addr: redisAddr}, asynq.Config{
		Concurrency: 1,
	})

	mux := asynq.NewServeMux()

	processor := worker.NewTaskProcessor(s3Client)

	mux.HandleFunc(models.TaskEncodeVideo, processor.HandleVideoEncodeTask)

	if err := asynqServer.Run(mux); err != nil {
		log.Fatalf("could not run transcoder worker: %v", err)
	}
}
