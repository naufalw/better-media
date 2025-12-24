package main

import (
	"better-media/internal/transcoder"
	"better-media/internal/uploader"
	"better-media/pkg/models"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	router := gin.Default()

	router.POST("/transcode", handleTranscode)

	port := os.Getenv("WORKER_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Worker listening on :%s", port)
	router.Run(":" + port)
}

func handleTranscode(c *gin.Context) {
	var req models.TranscodeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[%s] Received transcode request for video %s", req.JobID, req.VideoID)

	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})

	// do the transcoding
	go processTranscode(req)
}

func processTranscode(req models.TranscodeRequest) {

	log.Printf("[%s] Processing started...", req.JobID)

	payload := models.VideoEncodingPayload{
		VideoID:     req.VideoID,
		InputFile:   "", // we are using the url directly
		Resolutions: req.Resolutions,
	}

	pipeline, err := transcoder.NewEncodingPipeline(payload)
	if err != nil {
		log.Printf("[%s] Failed to create pipeline: %v", req.JobID, err)
		return
	}

	pipeline.Uploader = uploader.NewRemoteUploader(req.CallbackURL, req.JobID)
	pipeline.StreamURL = req.DownloadURL

	if err := pipeline.Run(context.Background(), nil); err != nil {
		pipeline.Uploader.NotifyFailed(err.Error())
		return
	}
	pipeline.Uploader.NotifyComplete()
}
