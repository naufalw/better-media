package api

import (
	"better-media/internal/api/middleware"
	_ "embed"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed dist/index.html
var indexHTML []byte

func (s *Server) registerRoutes() {
	v1 := s.Router.Group("/v1")

	// Public

	// Setup (this to setup the admin)
	v1.GET("/setup/status", s.handleSetupStatus)
	v1.POST("/setup/admin", s.handleSetupAdmin)

	// Auth
	v1.POST("/auth/register", s.handleRegister)
	v1.POST("/auth/login", s.handleLogin)

	// Playback (TEMPORARY)
	v1.GET("/videos/:videoId/playback/*assetPath", s.handlePlaybackProxy)
	v1.GET("/videos/:videoId/thumbnail/:size", s.handleGetThumbnail)
	v1.GET("/videos/:videoId/subtitles", s.handleGetSubtitles)
	v1.GET("/live/:streamKey/playback/*filePath", s.handleLivePlayback)

	// Worker Callbacks (TEMPORARY)
	v1.POST("/callbacks/:jobId/presign-upload", s.handlePresignUpload)
	v1.POST("/callbacks/:jobId/status", s.handleWorkerStatusUpdate)
	v1.POST("/callbacks/:jobId/progress", s.handleWorkerProgressUpdate)
	v1.POST("/callbacks/:jobId/metadata", s.handleWorkerMetadataUpdate)

	// PROTECTED
	protected := v1.Group("/")
	protected.Use(middleware.RequireAuth(s.JWT))
	{
		// Auth
		protected.GET("/auth/me", s.handleGetCurrentUser)

		// Videos
		protected.POST("/uploads", s.handleCreateUpload)
		protected.POST("/jobs/transcoding", s.handleCreateTranscodingJob)
		protected.GET("/jobs/:jobId", s.handleGetJobStatus)
		protected.GET("/videos", s.handleListVideos)
		protected.GET("/videos/:videoId", s.handleGetVideoDetails)
		protected.DELETE("/videos/:videoId", s.handleDeleteVideo)
		protected.PATCH("/videos/:videoId/library", s.handleMoveVideoToLibrary)

		// Livestream
		protected.GET("/live", s.handleListLiveStreams)
		protected.POST("/stream-keys", s.handleCreateStreamKey)
		protected.GET("/stream-keys", s.handleListStreamKeys)

		// API Keys
		protected.POST("/api-keys", s.handleCreateAPIKey)
		protected.GET("/api-keys", s.handleListAPIKeys)
		protected.DELETE("/api-keys/:id", s.handleDeleteAPIKey)

		// Libraries
		protected.POST("/libraries", s.handleCreateLibrary)
		protected.GET("/libraries", s.handleListLibraries)
		protected.GET("/libraries/:id", s.handleGetLibrary)
		protected.PUT("/libraries/:id", s.handleUpdateLibrary)
		protected.DELETE("/libraries/:id", s.handleDeleteLibrary)

		admin := protected.Group("/")
		admin.Use(middleware.RequireAdmin())
		{
			admin.GET("/users", s.handleListUsers)
			admin.PATCH("/users/:id/role", s.handleUpdateUserRole)
			admin.DELETE("/users/:id", s.handleDeleteUser)
		}

		protected.POST("/auth/change-password", s.handleChangePassword)
	}

	// Frontend
	s.Router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/v1") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		if strings.HasPrefix(path, "/admin") {
			c.Data(http.StatusOK, "text/html", indexHTML)
			return
		}

		if path == "/" {
			c.Redirect(http.StatusTemporaryRedirect, "/admin")
			return
		}

		c.String(http.StatusNotFound, "Page not found")
	})
}
