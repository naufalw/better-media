package main

import (
	"better-media/internal/storage"
	"better-media/pkg/models"
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
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
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "OPTIONS"}
	router.Use(cors.New(config))

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
		"url":       result.URL,
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
	info, err := api.AsynqClient.Enqueue(task, asynq.MaxRetry(0))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue task"})
		return
	}
	log.Printf("Enqueued task: id=%s queue=%s", info.ID, info.Queue)
	c.JSON(http.StatusOK, gin.H{"message": "Encoding job has been queued", "task_id": info.ID})
}

func (api *API) handleGetVideoDetails(c *gin.Context) {
	videoId := c.Param("videoId")

	appBaseURL := "http://localhost:8080"
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

	if !strings.HasSuffix(assetPath, ".m3u8") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset type"})
		return
	}

	keyInBucket := path.Join(videoId, strings.TrimPrefix(assetPath, "/"))

	playlistContent, err := api.S3Client.GetObject(c.Request.Context(), keyInBucket)
	if err != nil {
		log.Printf("!!! S3 GET FAILED !!! Key: [%s], Error: [%v]", keyInBucket, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}
	defer playlistContent.Close()

	var rewrittenPlaylist strings.Builder
	scanner := bufio.NewScanner(playlistContent)

	relativeDir := path.Dir(strings.TrimPrefix(assetPath, "/"))
	appBaseURL := "http://localhost:8080"
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			rewrittenPlaylist.WriteString(line + "\n")
			continue
		}

		var newURL string
		if strings.HasSuffix(line, ".m3u8") {
			nextAssetPath := path.Join("/", relativeDir, line)
			newURL = fmt.Sprintf("%s/v1/videos/%s/playback%s", appBaseURL, videoId, nextAssetPath)
		} else {
			segmentKey := path.Join(videoId, relativeDir, line)
			presignedURL, err := api.S3Client.GeneratePresignedGet(
				c.Request.Context(), segmentKey, 2*time.Second,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign URL"})
				return
			}
			newURL = presignedURL.URL
		}
		rewrittenPlaylist.WriteString(newURL + "\n")
	}

	if err := scanner.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read playlist"})
		return
	}

	c.Header("Cache-Control", "max-age=300")
	c.Data(http.StatusOK, "application/vnd.apple.mpegurl", []byte(rewrittenPlaylist.String()))
}
