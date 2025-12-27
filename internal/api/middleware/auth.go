package middleware

import (
	"better-media/internal/auth"
	"better-media/internal/database"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Middleware to check valid JWT
func RequireAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context for handlers
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// Check if user has admin role
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// require api key
func RequireAPIKey(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}
		key, err := db.ValidateAPIKey(apiKey)
		if err != nil || key == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		c.Set("api_key_id", key.ID)
		c.Set("api_key_scopes", key.Scopes)
		c.Next()
	}
}

// require necessary cope
func RequireScope(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopes, exists := c.Get("api_key_scopes")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No scopes found"})
			c.Abort()
			return
		}
		// Check if it exist
		scopeList := strings.Split(scopes.(string), ",")
		for _, s := range scopeList {
			if strings.TrimSpace(s) == scope || strings.TrimSpace(s) == "*" {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}
