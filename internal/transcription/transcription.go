package transcription

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Transcriber struct {
	Provider Provider
}

func NewTranscriber(provider Provider) *Transcriber {
	return &Transcriber{Provider: provider}
}

func (t *Transcriber) TranscribeVideo(ctx context.Context, videoPath, outputDir string) (string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// get audio
	audioPath := filepath.Join(outputDir, "audio.wav")
	log.Printf("Extracting audio to %s", audioPath)

	if err := extractAudio(ctx, videoPath, audioPath); err != nil {
		return "", fmt.Errorf("failed to extract audio: %w", err)
	}
	defer os.Remove(audioPath)

	// transcribe
	log.Printf("Sending audio to transcription API...")
	result, err := t.Provider.Transcribe(ctx, audioPath)
	if err != nil {
		return "", fmt.Errorf("transcription failed: %w", err)
	}
	log.Printf("Transcription complete: %d segments, language: %s", len(result.Segments), result.Language)

	// make VTT
	vttPath := filepath.Join(outputDir, "subtitles.vtt")
	if err := GenerateVTT(result, vttPath); err != nil {
		return "", fmt.Errorf("failed to generate VTT: %w", err)
	}
	log.Printf("Generated VTT file: %s", vttPath)
	return vttPath, nil
}

// Use FFmpeg to extract audio from video and convert to 16khz
func extractAudio(ctx context.Context, videoPath, audioPath string) error {
	args := []string{
		"-hide_banner", "-y",
		"-i", videoPath,
		"-vn",
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		audioPath,
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}
