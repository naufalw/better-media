package main

import (
	"better-media/internal/worker"
	"better-media/pkg/models"
	"log"

	"github.com/hibiken/asynq"
)

const redisAddr = "127.0.0.1:6379"

func main() {

	asynqServer := asynq.NewServer(asynq.RedisClientOpt{Addr: redisAddr}, asynq.Config{
		Concurrency: 1,
	})

	mux := asynq.NewServeMux()

	mux.HandleFunc(models.TaskEncodeVideo, worker.HandleVideoEncodeTask)

	if err := asynqServer.Run(mux); err != nil {
		log.Fatalf("could not run transcoder worker: %v", err)
	}
}
