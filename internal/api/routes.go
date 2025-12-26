package api

import (
	_ "embed"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed dist/index.html
var indexHTML []byte

// registerRoutes sets up all API routes
func (s *Server) registerRoutes() {
	v1 := s.Router.Group("/v1")
	{
		// Upload & Transcoding
		v1.POST("/uploads", s.handleCreateUpload)
		v1.POST("/jobs/transcoding", s.handleCreateTranscodingJob)
		v1.GET("/jobs/:jobId", s.handleGetJobStatus)

		// Videos
		v1.GET("/videos", s.handleListVideos)
		v1.GET("/videos/:videoId", s.handleGetVideoDetails)
		v1.DELETE("/videos/:videoId", s.handleDeleteVideo)
		v1.GET("/videos/:videoId/playback/*assetPath", s.handlePlaybackProxy)
		v1.GET("/videos/:videoId/thumbnail/:size", s.handleGetThumbnail)
		v1.GET("/videos/:videoId/subtitles", s.handleGetSubtitles)

		// Worker Callbacks (internal)
		v1.POST("/callbacks/:jobId/presign-upload", s.handlePresignUpload)
		v1.POST("/callbacks/:jobId/status", s.handleWorkerStatusUpdate)
		v1.POST("/callbacks/:jobId/progress", s.handleWorkerProgressUpdate)

		// Livestream
		v1.GET("/live", s.handleListLiveStreams)
		v1.GET("/live/:streamKey/playback/*filePath", s.handleLivePlayback)

		// Stream Keys
		v1.POST("/stream-keys", s.handleCreateStreamKey)
		v1.GET("/stream-keys", s.handleListStreamKeys)
	}

	// Front End
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
