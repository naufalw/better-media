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

	c.JSON(http.StatusOK, gin.H{
		"id":          library.ID,
		"name":        library.Name,
		"description": library.Description,
		"video_count": len(videos),
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
