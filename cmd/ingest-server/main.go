package main

import (
	"better-media/internal/storage"
	"better-media/pkg/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/hibiken/asynq"
)

type PresignedRequest struct {
	Title string `json:"title" binding:"required"`
	Id    string `json:"id" binding:"required"`
}

const redisAddr = "127.0.0.1:6379"

func main() {
	godotenv.Load()
	router := gin.Default()

	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer asynqClient.Close()

	router.POST("/ingest/init", func(c *gin.Context) {
		var req PresignedRequest

		err := c.BindJSON(&req)

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("received request text: %s", req.Title)
		log.Printf("received request id: %s", req.Id)

		s3Client, err := storage.NewS3Client(
			os.Getenv("S3_BUCKET_NAME"),
			os.Getenv("S3_ENDPOINT"),
			"auto",
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate presigned URL"})
			return
		}

		result, err := s3Client.GeneratePresignedPut(c, req.Id+"/"+req.Title)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate presigned URL"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"url":       result.URL,
			"expiresAt": time.Now().Add(time.Minute * 15).UnixMilli(),
		})
	})

	router.POST("/ingest/enqueue", func(c *gin.Context) {
		var req models.VideoEncodingPayload

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		task, err := models.NewVideoEncodingTask(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
			return
		}

		info, err := asynqClient.Enqueue(task, asynq.MaxRetry(3))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue task"})
			return
		}

		log.Printf("Enqueued task: id=%s queue=%s", info.ID, info.Queue)
		c.JSON(http.StatusOK, gin.H{"message": "Encoding job has been queued", "task_id": info.ID})
	})

	router.Run()
}
