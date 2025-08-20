package worker

import (
	"better-media/pkg/models"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type EncodingPipeline struct {
	Payload models.VideoEncodingPayload

	SourceInfo struct {
		Width    int
		Height   int
		HasAudio bool
	}

	TempDir            string
	DownloadedFilePath string
}

func NewEncodingPipeline(p models.VideoEncodingPayload) (*EncodingPipeline, error) {
	tempDir, err := os.MkdirTemp("", "media-*-"+p.VideoID)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	return &EncodingPipeline{
		Payload:            p,
		TempDir:            tempDir,
		DownloadedFilePath: filepath.Join(tempDir, p.InputFile),
	}, nil
}

func (p *EncodingPipeline) Run() error {
	log.Println("Stage: Run...")

	return nil
}

func (p *EncodingPipeline) Probe() error {
	log.Println("Stage: Probing...")
	return nil
}

func (p *EncodingPipeline) Encode() error {
	log.Println("Stage: Encoding...")
	return nil
}
