package api

import (
	"better-media/pkg/models"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PresignedRequest is the request body for creating an upload
type PresignedRequest struct {
	FileName string `json:"file_name" binding:"required"`
}

// Generates a presigned URL for uploading a video
func (s *Server) handleCreateUpload(c *gin.Context) {
	var req PresignedRequest
	videoID := uuid.New().String()

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	log.Printf("Creating upload for file: %s", req.FileName)

	objectKey := filepath.Join(videoID, "source", req.FileName)
	validDuration := time.Minute * 15

	result, err := s.Storage.GeneratePresignedPut(c.Request.Context(), objectKey, validDuration)
	if err != nil {
		log.Printf("Error generating presigned URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate presigned URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"videoId":   videoID,
		"url":       result.URL,
		"expiresAt": time.Now().Add(validDuration).UnixMilli(),
	})
}

// Start a new transcoding job
func (s *Server) handleCreateTranscodingJob(c *gin.Context) {
	var req models.VideoEncodingPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	jobID, err := s.Dispatcher.Dispatch(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Cannot encode: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Encoding job has been queued",
		"job_id":  jobID,
	})
}

// Return the status of a transcoding job
func (s *Server) handleGetJobStatus(c *gin.Context) {
	jobID := c.Param("jobId")

	job, err := s.DB.GetJob(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id":   job.ID,
		"video_id": job.VideoID,
		"status":   job.Status,
		"progress": job.Progress,
		"error":    job.Error,
	})
}
