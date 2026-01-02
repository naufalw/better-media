package api

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed dist/embed/index.html
var embedHTML string

func (s *Server) handleEmbed(c *gin.Context) {
	videoID := c.Param("videoId")

	video, err := s.DB.GetVideo(videoID)
	if err != nil || video == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	config := map[string]any{
		"src":         s.Config.BaseURL + "/v1/videos/" + videoID + "/playback/hls/master.m3u8",
		"poster":      s.Config.BaseURL + "/v1/videos/" + videoID + "/thumbnail/640",
		"subtitleUrl": s.Config.BaseURL + "/v1/videos/" + videoID + "/subtitles",
		"theme":       c.DefaultQuery("theme", "dark"),
		"autoPlay":    c.Query("autoplay") == "1",
	}

	configJSON, _ := json.Marshal(config)

	html := strings.Replace(embedHTML, "{ /* GO_INJECT_CONFIG */ }", string(configJSON), 1)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))

}
