package worker

import (
	"better-media/internal/storage"
	"better-media/pkg/models"
	"bytes"
	"context"
	"fmt"
	"io"
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
	objectKey := filepath.Join(p.Payload.VideoID, "source", p.Payload.InputFile)
	log.Printf("Attempting to download object: %s", objectKey)
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

	hlsOutputPath := filepath.Join(p.EncodedOutputPath, "hls")
	if err := os.MkdirAll(hlsOutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create hls output dir: %w", err)
	}

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

			renditionDir := filepath.Join(hlsOutputPath, renditionName)
			if err := os.MkdirAll(renditionDir, 0755); err != nil {
				return fmt.Errorf("failed to create rendition dir %s: %w", renditionDir, err)
			}

			cmd.Output(filepath.Join(renditionDir, "playlist.m3u8"))

			// Video codec and scaling is here, TODO: make this configurable
			cmd.Extra("-vf", fmt.Sprintf("scale=w=-2:h=%d", res)).
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

			/// HLS settings for this rendition
			cmd.Extra("-hls_time", "4").
				Extra("-hls_playlist_type", "vod").
				Extra("-hls_segment_filename", filepath.Join(renditionDir, renditionName+"_%03d.ts"))

			// Playlist for this rendition
			cmd.Extra(filepath.Join(renditionDir, "playlist.m3u8"))

			outputIndex++

		}

	}

	// Master playlist for VOD
	cmd.Extra("-f", "hls").
		Extra("-master_pl_name", "master.m3u8").
		Extra("-var_stream_map", strings.TrimSpace(streamMapBuilder.String()))

	cmd.Output(filepath.Join(hlsOutputPath, "master.m3u8"))

	log.Printf("[%s] Stage [3/5]: Executing FFmpeg command: %s\n", p.Payload.VideoID, cmd.String())

	buf := &bytes.Buffer{}

	cmd.PipeOutput(buf)

	err := cmd.RunWithContext(ctx)
	if err != nil {
		log.Printf("FFmpeg command failed: %v\nOutput: %s", err,
			buf.String())
		return fmt.Errorf("ffmpeg command failed: %w", err)
	}

	out, _ := io.ReadAll(buf)
	log.Printf("FFmpeg command output: %s", out)

	return err
}

func (p *EncodingPipeline) Upload(ctx context.Context, s3c *storage.S3Client) error {
	log.Printf("[%s] Stage [4/5]: Uploading to S3...\n", p.Payload.VideoID)

	return filepath.Walk(p.EncodedOutputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relativePath, err := filepath.Rel(p.EncodedOutputPath, path)
			if err != nil {
				return err
			}

			// 20 Aug: The object key is [videoId]/[hls]/[rendition]/file.ts
			objectKey := filepath.Join(p.Payload.VideoID, relativePath)

			log.Printf("Uploading %s to %s", path, objectKey)
			if err := s3c.UploadFile(ctx, objectKey, path); err != nil {
				return fmt.Errorf("failed to upload %s: %w", info.Name(), err)
			}
		}
		return nil
	})
}

func (p *EncodingPipeline) Cleanup() error {
	log.Println("Stage: Cleanup...")
	if err := os.RemoveAll(p.TempDir); err != nil {
		return fmt.Errorf("failed to remove temp dir %s: %w", p.TempDir, err)
	}
	return nil
}
