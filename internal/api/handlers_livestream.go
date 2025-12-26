package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Return all active live streams
func (s *Server) handleListLiveStreams(c *gin.Context) {
	if s.LivestreamManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "RTMP server not running"})
		return
	}

	streams := s.LivestreamManager.GetActiveStreams()
	streamInfos := make([]gin.H, 0, len(streams))

	for _, key := range streams {
		stream := s.LivestreamManager.GetStream(key)
		if stream != nil {
			streamInfos = append(streamInfos, gin.H{
				"key":        key,
				"started_at": stream.StartedAt,
				"hls_url":    fmt.Sprintf("/v1/live/%s/playback/index.m3u8", key),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"streams": streamInfos})
}

// Serves HLS files for live streams
func (s *Server) handleLivePlayback(c *gin.Context) {
	streamKey := c.Param("streamKey")
	filePath := c.Param("filePath")

	if len(filePath) > 0 && filePath[0] == '/' {
		filePath = filePath[1:]
	}

	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path"})
		return
	}

	fullPath := filepath.Join(s.Config.StreamsPath, streamKey, cleanPath)

	baseDir := filepath.Join(s.Config.StreamsPath, streamKey)
	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(baseDir)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path"})
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	contentType := "application/octet-stream"
	if strings.HasSuffix(filePath, ".m3u8") {
		contentType = "application/vnd.apple.mpegurl"
	} else if strings.HasSuffix(filePath, ".ts") {
		contentType = "video/mp2t"
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-cache")
	c.File(fullPath)
}

// Create a new stream key
func (s *Server) handleCreateStreamKey(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := s.DB.CreateStreamKey(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stream key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":   key.ID,
		"name": key.Name,
		"key":  key.Key,
	})
}

// Return all stream keys
func (s *Server) handleListStreamKeys(c *gin.Context) {
	keys, err := s.DB.ListStreamKeys()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list stream keys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stream_keys": keys})
}
