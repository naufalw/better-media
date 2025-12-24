package main

import (
	"better-media/internal/database"
	"better-media/internal/dispatcher"
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
)

type PresignedRequest struct {
	FileName string `json:"file_name" binding:"required"`
}

const redisAddr = "127.0.0.1:6379"

// main initializes application dependencies (environment, database, S3 client, and job dispatcher),
// registers HTTP routes for the API (uploads, transcoding jobs, job status, video details and playback),
// and starts the HTTP server on port 8080.
func main() {
	godotenv.Load()

	db, err := database.New("./data/katie.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "OPTIONS"}
	router.Use(cors.New(config))

	s3Client, err := storage.NewS3Client(
		os.Getenv("S3_BUCKET_NAME"),
		os.Getenv("S3_ENDPOINT"),
		"auto",
	)
	if err != nil {
		log.Fatalf("failed to create s3 client: %v", err)
	}

	var jobDispatcher dispatcher.JobDispatcher
	dispatcherMode := os.Getenv("DISPATCHER_MODE")
	if dispatcherMode == "http" {
		workerURL := os.Getenv("WORKER_URL")
		callbackURL := os.Getenv("CALLBACK_URL")
		if workerURL == "" || callbackURL == "" {
			log.Fatalf("WORKER_URL and CALLBACK_URL must be set when DISPATCHER_MODE=http")
		}
		jobDispatcher = dispatcher.NewHTTPDispatcher(s3Client, db, workerURL, callbackURL)
		log.Println("using external transcode")
	} else {
		jobDispatcher = dispatcher.NewLocalDispatcher(s3Client, db)
		log.Println("using local transcode")
	}
	api := &API{
		S3Client:   s3Client,
		Dispatcher: jobDispatcher,
		DB:         db,
	}

	// Version 1
	v1 := router.Group("/v1")
	{
		v1.POST("/uploads", api.handleCreateUpload)
		v1.POST("/jobs/transcoding", api.handleCreateTranscodingJob)
		v1.GET("/jobs/:jobId", api.handleGetJobStatus)

		v1.GET("/videos/:videoId", api.handleGetVideoDetails)
		v1.GET("/videos/:videoId/playback/*assetPath", api.handlePlaybackProxy)

		v1.POST("/callbacks/:jobId/presign-upload", api.handlePresignUpload)
		v1.POST("/callbacks/:jobId/status", api.handleWorkerStatusUpdate)
	}

	router.Run(":8080")
}

type API struct {
	S3Client   *storage.S3Client
	Dispatcher dispatcher.JobDispatcher
	DB         *database.DB
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
	jobID, err := api.Dispatcher.Dispatch(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Cannot encode: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Encoding job has been queued", "job_id": jobID})
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

	isMasterPlaylist := strings.HasSuffix(assetPath, "master.m3u8")

	if !strings.HasSuffix(assetPath, ".m3u8") && !strings.HasSuffix(assetPath, ".ts") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset type"})
		return
	}

	keyInBucket := path.Join(videoId, strings.TrimPrefix(assetPath, "/"))

	if strings.HasSuffix(assetPath, ".ts") {
		presignedURL, err := api.S3Client.GeneratePresignedGet(c.Request.Context(), keyInBucket, 1*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign segment URL"})
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, presignedURL.URL)
		return
	}

	playlistContent, err := api.S3Client.GetObject(c.Request.Context(), keyInBucket)
	if err != nil {
		log.Printf("!!! S3 GET FAILED !!! Key: [%s], Error: [%v]", keyInBucket, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}
	defer playlistContent.Close()

	if isMasterPlaylist {

		c.Header("Cache-Control", "max-age=2, must-revalidate")
		log.Printf("Serving master playlist for %s with short cache time.", videoId)
	} else {
		c.Header("Cache-Control", "max-age=3600")
	}

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

		nextAssetPath := path.Join("/", relativeDir, line)
		newURL := fmt.Sprintf("%s/v1/videos/%s/playback%s", appBaseURL, videoId, nextAssetPath)
		rewrittenPlaylist.WriteString(newURL + "\n")
	}

	if err := scanner.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read playlist"})
		return
	}

	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.String(http.StatusOK, rewrittenPlaylist.String())
}

func (api *API) handleGetJobStatus(c *gin.Context) {
	jobID := c.Param("jobId")

	job, err := api.DB.GetJob(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id":   job.ID,
		"video_id": job.VideoID,
		"status":   job.Status,
		"error":    job.Error,
	})
}

// from transcoder to request presigned url for uploading
func (api *API) handlePresignUpload(c *gin.Context) {

	var req struct {
		Files []string `json:"files"` // key, such as ["<videoID"/hls/360p/playlist.m3u8"]
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	urls := make(map[string]string)

	for _, file := range req.Files {
		presigned, err := api.S3Client.GeneratePresignedPut(c.Request.Context(), file, time.Minute*15)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate presign for %s", file)})
			return
		}

		urls[file] = presigned.URL

	}

	c.JSON(http.StatusOK, gin.H{"urls": urls})

}

// from transcoder to main api to update status
func (api *API) handleWorkerStatusUpdate(c *gin.Context) {
	jobID := c.Param("jobId")

	var req struct {
		Status string  `json:"status"` // "playable", "completed", "failed"
		Error  *string `json:"error"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := api.DB.UpdateJobStatus(jobID, req.Status, req.Error); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	log.Printf("[%s] Worker reported status: %s", jobID, req.Status)
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}
