package worker

import (
	"better-media/internal/storage"
	"better-media/pkg/models"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bitcodr/gompeg"
	fluentffmpeg "github.com/modfy/fluent-ffmpeg"
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

	if err := p.Download(ctx, s3c); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	if err := p.Probe(); err != nil {
		return fmt.Errorf("failed to probe file: %w", err)
	}
	if err := p.Encode(ctx); err != nil {
		return fmt.Errorf("failed to encode file: %w", err)
	}

	return nil
}

func (p *EncodingPipeline) Download(ctx context.Context, s3c *storage.S3Client) error {
	log.Printf("[%s] Stage [1/5]: Downloading from S3...\n", p.Payload.VideoID)
	objectKey := filepath.Join(p.Payload.VideoID, p.Payload.InputFile)
	return s3c.DownloadFile(ctx, objectKey, p.DownloadedFilePath)
}

func (p *EncodingPipeline) Probe() error {
	log.Printf("[%s] Stage [2/5]: Probing input file...\n", p.Payload.VideoID)

	data, err := fluentffmpeg.Probe(p.DownloadedFilePath)

	if err != nil {
		return fmt.Errorf("fluentffmpeg.Probe failed: %w", err)
	}

	streams, ok := data["streams"].([]any)
	if !ok {
		return fmt.Errorf("could not find streams in ffprobe output")
	}

	foundVideo := false

	for _, streamData := range streams {
		stream, ok := streamData.(map[string]any)
		if !ok {
			continue
		}

		codecType, ok := stream["codec_type"].(string)

		switch codecType {
		case "video":
			if width, ok := stream["width"].(float64); ok {
				p.SourceInfo.Width = int(width)
			}
			if height, ok := stream["height"].(float64); ok {
				p.SourceInfo.Height = int(height)
			}
			foundVideo = true
		case "audio":
			p.SourceInfo.HasAudio = true
		}
	}

	if !foundVideo {
		return fmt.Errorf("no video stream found in file")
	}

	log.Printf("Probe complete. Resolution: %dx%d, HasAudio: %t", p.SourceInfo.Width, p.SourceInfo.Height, p.SourceInfo.HasAudio)
	return nil

}

func (p *EncodingPipeline) Encode(ctx context.Context) error {
	log.Printf("[%s] Stage [3/5]: Encoding...\n", p.Payload.VideoID)

	cmd := gompeg.New().
		Input(p.DownloadedFilePath).
		Extra("-sc_threshold", "0").
		Extra("-g", "48").
		Extra("-keyint_min", "48")

	// Building rendition (scaled down) streams
	var streamMapBuilder strings.Builder
	var outputIndex int = 0

	for _, res := range p.Payload.Resolutions {
		if res <= p.SourceInfo.Height {
			renditionName := strconv.Itoa(res) + "p"

			// Video codec and scaling is here, TODO: make this configurable
			cmd.Extra("-vf", fmt.Sprintf("scale=2=-2:h=%d", res)).
				Extra("-c:v", "h264").
				Extra("-profile:v", "main").
				Extra("-crf", "23")
			if p.SourceInfo.HasAudio {

				// Audio codec and bitrate is here, TODO: make this configurable
				cmd.Extra("-c:a", "aac").
					Extra("-b:a", "128k")

				cmd.Extra("-map", "0:v:0", "-map", "0:a:0")
				fmt.Fprintf(&streamMapBuilder, "v:%d,a:%d,name:%s ", outputIndex, outputIndex, renditionName)
			} else {
				// Here we only map video
				cmd.Extra("-map", "0:v:0")
				fmt.Fprintf(&streamMapBuilder, "v:%d,name:%s ", outputIndex, renditionName)
			}

			/// HLS
			cmd.Extra("-hls_time", "4").
				Extra("-hls_playlist_type", "vod").
				Extra("-hls_segment_filename", filepath.Join(p.EncodedOutputPath, renditionName+"_%03d.ts"))

			// Playlist for this rendition
			cmd.Extra(filepath.Join(p.EncodedOutputPath, renditionName+".m3u8"))

			outputIndex++

		}

	}

	// Master playlist for VOD
	cmd.Extra("-f", "hls").
		Extra("-master_pl_name", "master.m3u8").
		Extra("-var_stream_map", strings.TrimSpace(streamMapBuilder.String()))

	cmd.Output(p.EncodedOutputPath)

	log.Printf("[%s] Stage [3/5]: Executing FFmpeg command: %s\n", p.Payload.VideoID, cmd.String())
	return cmd.RunWithContext(ctx)
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
