package main

import (
	"better-media/pkg/models"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/hibiken/asynq"
)

type PresignedRequest struct {
	Title string `json:"title" binding:"required"`
	Id    string `json:"id" binding:"required"`
}

type WorkEncodeRequest struct {
	Id string `json:"id" binding:"required"`
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

		url, expire := generatePresigned(c, req.Title, req.Id)

		c.JSON(http.StatusOK, gin.H{
			"url":       url,
			"expiresAt": expire.UnixMilli(),
		})
	})

	router.POST("/ingest/enqueue", func(c *gin.Context) {
		var req models.VideoEncodingPayload

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		task, err := models.NewVideoEncodingTask(req.VideoID)
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

func generatePresigned(c *gin.Context, title string, id string) (string, time.Time) {
	accessKeyId, accessKeySecret := os.Getenv("S3_ACCESS_KEY_ID"), os.Getenv("S3_ACCESS_KEY_SECRET")
	accountId := os.Getenv("S3_ACCOUNT_ID")
	bucketName := os.Getenv("S3_BUCKET_NAME")
	cfg, err := config.LoadDefaultConfig(c, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")), config.WithRegion("auto"))

	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})

	presignClient := s3.NewPresignClient(client)
	presignResult, err := presignClient.PresignPutObject(c, &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    aws.String(id + "/" + title),
	}, s3.WithPresignExpires(time.Minute*15))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate presigned URL",
		})
	}

	return presignResult.URL, time.Now().Add(time.Minute * 15)

}
