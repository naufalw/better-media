// This is a temporary utility to choose video and audio bitrates based on the height of the video.
// This will be removed ASAP along with more adaptive settings using ffprobe and user's preferences.

package transcoder

import (
	"os/exec"
	"strings"
)

func chooseVideoBitrate(h int) string {
	switch {
	case h >= 1080:
		return "5000k"
	case h >= 720:
		return "2800k"
	case h >= 480:
		return "1400k"
	case h >= 360:
		return "800k"
	default:
		return "400k"
	}
}

func getBandwidthForHeight(h int) int {
	switch {
	case h >= 1080:
		return 5000000
	case h >= 720:
		return 2800000
	case h >= 480:
		return 1400000
	case h >= 360:
		return 800000
	default:
		return 400000
	}
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
