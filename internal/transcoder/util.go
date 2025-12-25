// This is a temporary utility to choose video and audio bitrates based on the height of the video.
// This will be removed ASAP along with more adaptive settings using ffprobe and user's preferences.

package transcoder

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func calculateRenditionBandwidth(renditionDir string, durationSec float64) int {
	if durationSec <= 0 {
		return 0
	}

	var totalBytes int64
	filepath.Walk(renditionDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalBytes += info.Size()
		}
		return nil
	})

	// Bandwidth = (bytes * 8 bits) / duration in seconds
	return int(float64(totalBytes*8) / durationSec)
}

func chooseAudioBitrate(h int) string {
	if h >= 720 {
		return "128k"
	}
	return "96k"
}

var cachedEncoder string

func getBestVideoEncoder() string {
	if cachedEncoder != "" {
		return cachedEncoder
	}

	out, err := exec.Command("ffmpeg", "-encoders").CombinedOutput()

	if err != nil {
		cachedEncoder = "libx264" // ffmpeg is missing or error
		return cachedEncoder
	}

	s := string(out)
	if strings.Contains(s, "h264_videotoolbox") {
		cachedEncoder = "h264_videotoolbox"
	} else if strings.Contains(s, "h264_nvenc") {
		cachedEncoder = "h264_nvenc"
	} else {
		cachedEncoder = "libx264"
	}

	return cachedEncoder

}

// Get the quality control flags to handle different encoder
func getQualityArgs() []string {
	encoder := getBestVideoEncoder()

	switch encoder {
	case "h264_videotoolbox":
		// VideoToolbox -q:v (0-100, higher = better)
		return []string{"-q:v", "65"}
	case "h264_nvenc":
		// NVENC -cq (0-51, lower = better)
		return []string{"-cq", "23"}
	default:
		// libx264 -crf (0-51, lower = better)
		return []string{"-crf", "23"}
	}
}

func selectResolutions(sourceHeight int) []int {

	standardResolutions := []int{360, 480, 720, 1080, 1440, 2160}

	var selected []int
	for _, res := range standardResolutions {
		if res <= sourceHeight {
			selected = append(selected, res)
		}
	}

	included := false
	for _, res := range selected {
		if res == sourceHeight {
			included = true
			break
		}
	}
	if !included {
		selected = append(selected, sourceHeight)
	}

	return selected
}

// read ffmpeg stderr and call onProgress with current time in seconds
func parseFFmpegProgress(r io.Reader, onProgress func(currentSeconds float64)) {
	scanner := bufio.NewScanner(r)
	timeRegex := regexp.MustCompile(`time=(\d+):(\d+):(\d+)\.(\d+)`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := timeRegex.FindStringSubmatch(line)
		if len(matches) == 5 {
			hours, _ := strconv.Atoi(matches[1])
			mins, _ := strconv.Atoi(matches[2])
			secs, _ := strconv.Atoi(matches[3])
			ms, _ := strconv.Atoi(matches[4])

			currentSeconds := float64(hours)*3600 + float64(mins)*60 + float64(secs) + float64(ms)/100
			log.Printf("Progress parsed: %.2f seconds", currentSeconds)
			onProgress(currentSeconds)
		}
	}
}
