package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// POST /v1/libraries
func (s *Server) handleCreateLibrary(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	library, err := s.DB.CreateLibrary(req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create library"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          library.ID,
		"name":        library.Name,
		"description": library.Description,
		"created_at":  library.CreatedAt,
	})
}

// GET /v1/libraries
func (s *Server) handleListLibraries(c *gin.Context) {
	libraries, err := s.DB.ListLibraries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list libraries"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"libraries": libraries})
}

// GET /v1/libraries/:id
func (s *Server) handleGetLibrary(c *gin.Context) {
	id := c.Param("id")

	library, err := s.DB.GetLibrary(id)
	if err != nil || library == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Library not found"})
		return
	}

	videos, _ := s.DB.ListVideosByLibrary(id)

	videoResponses := make([]gin.H, 0, len(videos))
	for _, v := range videos {
		videoResponses = append(videoResponses, gin.H{
			"id":                v.ID,
			"library_id":        v.LibraryID,
			"title":             v.Title,
			"description":       v.Description,
			"status":            v.Status,
			"source":            v.Source,
			"duration_ms":       v.DurationMs,
			"file_size_bytes":   v.FileSizeBytes,
			"resolution_width":  v.ResolutionWidth,
			"resolution_height": v.ResolutionHeight,
			"created_at":        v.CreatedAt,
			"updated_at":        v.UpdatedAt,
			"playback_url":      "/v1/videos/" + v.ID + "/playback/hls/master.m3u8",
			"thumbnail_url":     "/v1/videos/" + v.ID + "/thumbnail/320",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          library.ID,
		"name":        library.Name,
		"description": library.Description,
		"video_count": len(videos),
		"videos":      videoResponses,
		"created_at":  library.CreatedAt,
		"updated_at":  library.UpdatedAt,
	})
}

// PUT /v1/libraries/:id
func (s *Server) handleUpdateLibrary(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.DB.UpdateLibrary(id, req.Name, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update library"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Library updated"})
}

// DELETE /v1/libraries/:id
func (s *Server) handleDeleteLibrary(c *gin.Context) {
	id := c.Param("id")

	if err := s.DB.DeleteLibrary(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete library"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Library deleted"})
}
