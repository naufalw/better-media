package transcription

import (
	"fmt"
	"os"
	"strings"
)

// Creates WebVTT file from transcription segments
func GenerateVTT(result *TranscriptionResult, outputPath string) error {
	var sb strings.Builder

	// VTT header
	sb.WriteString("WEBVTT\n")

	sb.WriteString(fmt.Sprintf("Language: %s\n\n", result.Language))

	for i, seg := range result.Segments {
		// Cue identifier
		sb.WriteString(fmt.Sprintf("%d\n", i+1))
		// Timestamps: 00:00:00.000 --> 00:00:05.000
		sb.WriteString(fmt.Sprintf("%s --> %s\n", formatVTTTime(seg.Start), formatVTTTime(seg.End)))
		// Text
		sb.WriteString(strings.TrimSpace(seg.Text))
		sb.WriteString("\n\n")
	}
	return os.WriteFile(outputPath, []byte(sb.String()), 0o644)
}

// Converts seconds to VTT timestamp format: HH:MM:SS.mmm
func formatVTTTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, millis)
}
