package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// POST /v1/api-keys
func (s *Server) handleCreateAPIKey(c *gin.Context) {
	var req struct {
		Name      string  `json:"name" binding:"required"`
		Scopes    string  `json:"scopes"` // e.g., "videos:read,videos:write"
		ExpiresAt *string `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Parse expires_at if provided
	key, plainKey, err := s.DB.CreateAPIKey(req.Name, req.Scopes, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":         key.ID,
		"name":       key.Name,
		"key":        plainKey,
		"key_prefix": key.KeyPrefix,
		"scopes":     key.Scopes,
		"message":    "one time",
	})
}

// GET /v1/api-keys
func (s *Server) handleListAPIKeys(c *gin.Context) {
	keys, err := s.DB.ListAPIKeys()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list API keys"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"api_keys": keys})
}

// DELETE /v1/api-keys/:id
func (s *Server) handleDeleteAPIKey(c *gin.Context) {
	id := c.Param("id")
	if err := s.DB.DeleteAPIKey(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "API key deleted"})
}
