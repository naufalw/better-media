package api

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Returns all videos
func (s *Server) handleListVideos(c *gin.Context) {
	videos, err := s.DB.ListVideos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list videos"})
		return
	}

	result := make([]gin.H, 0, len(videos))
	for _, v := range videos {
		result = append(result, gin.H{
			"id":            v.ID,
			"title":         v.Title,
			"status":        v.Status,
			"source":        v.Source,
			"created_at":    v.CreatedAt,
			"playback_url":  fmt.Sprintf("/v1/videos/%s/playback/hls/master.m3u8", v.ID),
			"thumbnail_url": fmt.Sprintf("/v1/videos/%s/thumbnail/320", v.ID),
		})
	}

	c.JSON(http.StatusOK, gin.H{"videos": result})
}

// Returns details for a specific video
func (s *Server) handleGetVideoDetails(c *gin.Context) {
	videoID := c.Param("videoId")

	video, err := s.DB.GetVideo(videoID)
	if err != nil || video == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	playbackURL := fmt.Sprintf("/v1/videos/%s/playback/hls/master.m3u8", videoID)
	thumbnailURL := fmt.Sprintf("/v1/videos/%s/thumbnail/320", videoID)
	subtitleURL := fmt.Sprintf("/v1/videos/%s/subtitles", videoID)

	progress := 0
	job, err := s.DB.GetLatestJobByVideoID(videoID)
	if err == nil && job != nil {
		progress = job.Progress
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                video.ID,
		"library_id":        video.LibraryID,
		"title":             video.Title,
		"description":       video.Description,
		"status":            video.Status,
		"source":            video.Source,
		"duration_ms":       video.DurationMs,
		"file_size_bytes":   video.FileSizeBytes,
		"resolution_width":  video.ResolutionWidth,
		"resolution_height": video.ResolutionHeight,
		"created_at":        video.CreatedAt,
		"playback_url":      playbackURL,
		"thumbnail_url":     thumbnailURL,
		"subtitle_url":      subtitleURL,
		"progress":          progress,
	})
}

// Deletes a video and its assets
func (s *Server) handleDeleteVideo(c *gin.Context) {
	videoID := c.Param("videoId")

	video, err := s.DB.GetVideo(videoID)
	if err != nil || video == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	if err := s.Storage.DeletePrefix(c.Request.Context(), videoID+"/"); err != nil {
		log.Printf("Warning: failed to delete S3 files for %s: %v", videoID, err)
	}

	if err := s.DB.DeleteVideo(videoID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video"})
		return
	}

	log.Printf("[DELETE] Video %s deleted", videoID)
	c.JSON(http.StatusOK, gin.H{"message": "Video deleted", "video_id": videoID})
}

// handlePlaybackProxy serves HLS playlists and redirects segments to S3
func (s *Server) handlePlaybackProxy(c *gin.Context) {
	videoID := c.Param("videoId")
	assetPath := c.Param("assetPath")

	isMasterPlaylist := strings.HasSuffix(assetPath, "master.m3u8")

	if !strings.HasSuffix(assetPath, ".m3u8") && !strings.HasSuffix(assetPath, ".ts") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset type"})
		return
	}

	keyInBucket := path.Join(videoID, strings.TrimPrefix(assetPath, "/"))

	// For .ts segments, redirect to presigned S3 URL
	if strings.HasSuffix(assetPath, ".ts") {
		presignedURL, err := s.Storage.GeneratePresignedGet(c.Request.Context(), keyInBucket, 1*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign segment URL"})
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, presignedURL.URL)
		return
	}

	// For .m3u8 playlists, rewrite URLs
	playlistContent, err := s.Storage.GetObject(c.Request.Context(), keyInBucket)
	if err != nil {
		log.Printf("S3 GET failed for key [%s]: %v", keyInBucket, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}
	defer playlistContent.Close()

	if isMasterPlaylist {
		c.Header("Cache-Control", "max-age=2, must-revalidate")
	} else {
		c.Header("Cache-Control", "max-age=3600")
	}

	var rewrittenPlaylist strings.Builder
	scanner := bufio.NewScanner(playlistContent)
	relativeDir := path.Dir(strings.TrimPrefix(assetPath, "/"))

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			rewrittenPlaylist.WriteString(line + "\n")
			continue
		}

		nextAssetPath := path.Join("/", relativeDir, line)
		newURL := fmt.Sprintf("%s/v1/videos/%s/playback%s", s.Config.BaseURL, videoID, nextAssetPath)
		rewrittenPlaylist.WriteString(newURL + "\n")
	}

	if err := scanner.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read playlist"})
		return
	}

	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.String(http.StatusOK, rewrittenPlaylist.String())
}

// Returns a video thumbnail
func (s *Server) handleGetThumbnail(c *gin.Context) {
	videoID := c.Param("videoId")
	size := c.Param("size")

	validSizes := map[string]bool{"320": true, "640": true, "1280": true}
	if !validSizes[size] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid size. Must be one of: 320, 640, 1280"})
		return
	}

	objectKey := fmt.Sprintf("%s/thumbnails/thumb_%s.jpg", videoID, size)

	presigned, err := s.Storage.GeneratePresignedGet(c.Request.Context(), objectKey, time.Hour)
	if err != nil {
		log.Printf("Error generating presigned URL for thumbnail %s: %v", objectKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve thumbnail"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, presigned.URL)
}

// Returns video subtitles
func (s *Server) handleGetSubtitles(c *gin.Context) {
	videoID := c.Param("videoId")
	objectKey := fmt.Sprintf("%s/subtitles/subtitles.vtt", videoID)

	presigned, err := s.Storage.GeneratePresignedGet(c.Request.Context(), objectKey, time.Hour)
	if err != nil {
		log.Printf("Error generating presigned URL for subtitles %s: %v", objectKey, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Subtitles not found"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, presigned.URL)
}
