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
	FileName  string  `json:"file_name" binding:"required"`
	Title     string  `json:"title"`
	LibraryID *string `json:"library_id"`
}

// Generates a presigned URL for uploading a video
func (s *Server) handleCreateUpload(c *gin.Context) {
	var req PresignedRequest
	videoID := uuid.New().String()

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	title := req.Title
	if title == "" {
		title = req.FileName
	}

	log.Printf("Creating upload for file: %s, library: %v", req.FileName, req.LibraryID)

	objectKey := filepath.Join(videoID, "source", req.FileName)
	validDuration := time.Minute * 15

	result, err := s.Storage.GeneratePresignedPut(c.Request.Context(), objectKey, validDuration)
	if err != nil {
		log.Printf("Error generating presigned URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate presigned URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"videoId":    videoID,
		"url":        result.URL,
		"expiresAt":  time.Now().Add(validDuration).UnixMilli(),
		"title":      title,
		"library_id": req.LibraryID,
	})
}

// Start a new transcoding job
func (s *Server) handleCreateTranscodingJob(c *gin.Context) {

	var req struct {
		VideoID     string  `json:"video_id" binding:"required"`
		InputFile   string  `json:"input_file" binding:"required"`
		Title       string  `json:"title"`
		LibraryID   *string `json:"library_id"`
		Resolutions []int   `json:"resolutions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	title := req.Title
	if title == "" {
		title = req.InputFile
	}
	s.DB.CreateVideo(req.VideoID, title, "upload", req.LibraryID)
	payload := models.VideoEncodingPayload{
		VideoID:     req.VideoID,
		InputFile:   req.InputFile,
		Resolutions: req.Resolutions,
	}

	jobID, err := s.Dispatcher.Dispatch(c, payload)
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

// PATCH /v1/videos/:videoId/library
func (s *Server) handleMoveVideoToLibrary(c *gin.Context) {
	videoID := c.Param("videoId")
	var req struct {
		LibraryID *string `json:"library_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.DB.UpdateVideoLibrary(videoID, req.LibraryID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update video"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Video library updated"})
}
