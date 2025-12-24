package transcoder

import (
	"better-media/pkg/models"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	router := gin.Default()

	router.POST("/transcode", handleTranscode)

	port := os.Getenv("WORKER_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Worker listening on :%s", port)
	router.Run(":" + port)
}

func handleTranscode(c *gin.Context) {
	var req models.TranscodeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[%s] Received transcode request for video %s", req.JobID, req.VideoID)

	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})

	// do the transcoding
	go processTranscode(req)
}

func processTranscode(req models.TranscodeRequest) {
	// TODO: Implement
	// 1. Download source from req.DownloadURL
	// 2. Run encoding pipeline
	// 3. Request presigned PUT URLs via req.CallbackURL
	// 4. Upload files
	// 5. Report completion

	log.Printf("[%s] Processing started...", req.JobID)
}
