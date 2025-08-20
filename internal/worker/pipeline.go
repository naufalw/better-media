package worker

import (
	"better-media/internal/storage"
	"better-media/pkg/models"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type FFProbeStream struct {
	CodecType string `json:"codec_type"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

type EncodingPipeline struct {
	Payload models.VideoEncodingPayload

	SourceInfo struct {
		Width    int
		Height   int
		HasAudio bool
	}

	TempDir            string
	DownloadedFilePath string
	EncodedOutputPath  string
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
		EncodedOutputPath:  filepath.Join(tempDir, "encoded"),
	}, nil
}

func (p *EncodingPipeline) Run(ctx context.Context, s3c *storage.S3Client) error {
	log.Println("Stage: Run...")

	defer p.Cleanup()

	if err := p.Download(); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	if err := p.Probe(); err != nil {
		return fmt.Errorf("failed to probe file: %w", err)
	}
	if err := p.Encode(); err != nil {
		return fmt.Errorf("failed to encode file: %w", err)
	}

	return nil
}

func (p *EncodingPipeline) Download() error {

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

func (p *EncodingPipeline) Upload() error {
	log.Println("Stage: Uploading...")
	return nil
}

func (p *EncodingPipeline) Cleanup() error {
	log.Println("Stage: Cleanup...")
	if err := os.RemoveAll(p.TempDir); err != nil {
		return fmt.Errorf("failed to remove temp dir %s: %w", p.TempDir, err)
	}
	return nil
}
