// This is a temporary utility to choose video and audio bitrates based on the height of the video.
// This will be removed ASAP along with more adaptive settings using ffprobe and user's preferences.

package worker

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
