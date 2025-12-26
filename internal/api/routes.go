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

		// Livestream
		protected.GET("/live", s.handleListLiveStreams)
		protected.POST("/stream-keys", s.handleCreateStreamKey)
		protected.GET("/stream-keys", s.handleListStreamKeys)
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
