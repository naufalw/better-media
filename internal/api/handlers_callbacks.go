package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Generate presigned URLs for worker uploads
func (s *Server) handlePresignUpload(c *gin.Context) {
	var req struct {
		Files []string `json:"files"` // e.g., ["<videoID>/hls/360p/playlist.m3u8"]
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	urls := make(map[string]string)

	for _, file := range req.Files {
		presigned, err := s.Storage.GeneratePresignedPut(c.Request.Context(), file, time.Minute*15)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to generate presign for %s", file),
			})
			return
		}
		urls[file] = presigned.URL
	}

	c.JSON(http.StatusOK, gin.H{"urls": urls})
}

// Receives status updates from transcoding workers
func (s *Server) handleWorkerStatusUpdate(c *gin.Context) {
	jobID := c.Param("jobId")

	var req struct {
		Status string  `json:"status"` // "playable", "completed", "failed"
		Error  *string `json:"error"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.DB.UpdateJobStatus(jobID, req.Status, req.Error); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	job, err := s.DB.GetJob(jobID)
	if err == nil && job != nil {
		switch req.Status {
		case "completed":
			s.DB.UpdateVideoStatus(job.VideoID, "ready")
		case "failed":
			s.DB.UpdateVideoStatus(job.VideoID, "failed")
		}
	}

	log.Printf("[%s] Worker reported status: %s", jobID, req.Status)
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// Rceives progress updates from transcoding workers -> percentage of completion
func (s *Server) handleWorkerProgressUpdate(c *gin.Context) {
	jobID := c.Param("jobId")

	var req struct {
		Progress int `json:"progress"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.DB.UpdateJobProgress(jobID, req.Progress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// metadata update from workers
func (s *Server) handleWorkerMetadataUpdate(c *gin.Context) {

	var req struct {
		VideoID          string `json:"video_id"`
		DurationMs       int64  `json:"duration_ms"`
		FileSizeBytes    int64  `json:"file_size_bytes"`
		ResolutionWidth  int    `json:"resolution_width"`
		ResolutionHeight int    `json:"resolution_height"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.DB.UpdateVideoMetadata(req.VideoID, req.DurationMs, req.FileSizeBytes, req.ResolutionWidth, req.ResolutionHeight); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update video metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}
