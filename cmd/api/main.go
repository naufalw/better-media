package main

import (
	"better-media/internal/storage"
	"better-media/pkg/models"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/hibiken/asynq"
)

type PresignedRequest struct {
	FileName string `json:"file_name" binding:"required"`
}

const redisAddr = "127.0.0.1:6379"

func main() {
	godotenv.Load()
	router := gin.Default()

	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer asynqClient.Close()

	s3Client, err := storage.NewS3Client(
		os.Getenv("S3_BUCKET_NAME"),
		os.Getenv("S3_ENDPOINT"),
		"auto",
	)
	if err != nil {
		log.Fatalf("failed to create s3 client: %v", err)
	}

	api := &API{
		S3Client:    s3Client,
		AsynqClient: asynqClient,
	}

	// Version 1
	v1 := router.Group("/v1")
	{
		v1.POST("/uploads", api.handleCreateUpload)
		v1.POST("/jobs/transcoding", api.handleCreateTranscodingJob)

		v1.GET("/videos/:videoId", api.handleGetVideoDetails)
		v1.GET("/videos/:videoId/playback/*assetPath", api.handlePlaybackProxy)
	}

	router.Run()
}

type API struct {
	S3Client    *storage.S3Client
	AsynqClient *asynq.Client
}

func (api *API) handleCreateUpload(c *gin.Context) {
	var req PresignedRequest
	videoId := uuid.New().String()

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	log.Printf("received request text: %s", req.FileName)

	objectKey := filepath.Join(videoId, "source", req.FileName)

	validDuration := time.Minute * 15

	result, err := api.S3Client.GeneratePresignedPut(c.Request.Context(), objectKey, validDuration)
	if err != nil {
		log.Printf("Error generating presigned URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate presigned URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"videoId":   videoId,
		"url":       result,
		"expiresAt": time.Now().Add(validDuration).UnixMilli(),
	})
}

func (api *API) handleCreateTranscodingJob(c *gin.Context) {
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
	info, err := api.AsynqClient.Enqueue(task, asynq.MaxRetry(3))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue task"})
		return
	}
	log.Printf("Enqueued task: id=%s queue=%s", info.ID, info.Queue)
	c.JSON(http.StatusOK, gin.H{"message": "Encoding job has been queued", "task_id": info.ID})
}

func (api *API) handleGetVideoDetails(c *gin.Context) {
	videoId := c.Param("videoId")

	appBaseURL := os.Getenv("APP_BASE_URL")
	playbackUrl := fmt.Sprintf("%s/v1/videos/%s/playback/hls/master.m3u8", appBaseURL, videoId)

	c.JSON(http.StatusOK, gin.H{
		"videoId":     videoId,
		"status":      "PROCESSED",        // mock
		"title":       "My Awesome Video", // mock
		"playbackUrl": playbackUrl,
	})
}

func (api *API) handlePlaybackProxy(c *gin.Context) {
	videoId := c.Param("videoId")
	assetPath := c.Param("assetPath")

	objectKey := filepath.Join(videoId, strings.TrimPrefix(assetPath, "/"))

	validDuration := time.Second * 30

	presignedRequest, err := api.S3Client.GeneratePresignedGet(c.Request.Context(), objectKey, validDuration)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset not found"})
		return
	}

	c.Redirect(http.StatusFound, presignedRequest.URL)
}
