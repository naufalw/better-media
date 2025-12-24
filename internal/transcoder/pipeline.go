package transcoder

import (
	"better-media/internal/storage"
	"better-media/internal/uploader"
	"better-media/pkg/models"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

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

	StreamURL string

	// to unify the upolading process between monolith and distributed
	Uploader uploader.Uploader
}

// NewEncodingPipeline creates a temporary working directory for the given VideoEncodingPayload and returns
// an initialized EncodingPipeline configured to use that directory.
// - TempDir set to the created directory
// - DownloadedFilePath set to "<TempDir>/<InputFile>"
// - EncodedOutputPath set to "<TempDir>/encoded".
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

	objectKey := filepath.Join(p.Payload.VideoID, "source", p.Payload.InputFile)
	presignedGet, err := s3c.GeneratePresignedGet(ctx, objectKey, time.Hour)
	if err != nil {
		return fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	p.StreamURL = presignedGet.URL

	defer p.Cleanup()

	if err := p.Probe(); err != nil {
		return fmt.Errorf("failed to probe file: %w", err)
	}
	if err := p.Encode(ctx, s3c); err != nil {
		return fmt.Errorf("failed to encode file: %w", err)
	}

	if err := p.Upload(ctx); err != nil {
		return fmt.Errorf("failed to upload encoded files: %w", err)
	}

	log.Printf("[%s] Encoding pipeline completed successfully.\n", p.Payload.VideoID)

	return nil
}

func (p *EncodingPipeline) Download(ctx context.Context, s3c *storage.S3Client) error {
	log.Printf("[%s] Stage [1/5]: Downloading from S3...\n", p.Payload.VideoID)
	objectKey := filepath.Join(p.Payload.VideoID, "source", p.Payload.InputFile)
	log.Printf("[%s] Attempting to download object: %s", p.Payload.VideoID, objectKey)
	return s3c.DownloadFile(ctx, objectKey, p.DownloadedFilePath)
}

func (p *EncodingPipeline) Probe() error {
	log.Printf("[%s] Stage [2/5]: Probing input file...\n", p.Payload.VideoID)

	data, err := fluentffmpeg.Probe(p.StreamURL)

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

	log.Printf("[%s] Probe complete. Resolution: %dx%d, HasAudio: %t", p.Payload.VideoID,
		p.SourceInfo.Width, p.SourceInfo.Height, p.SourceInfo.HasAudio)
	return nil

}

type completedRendition struct {
	Width        int
	Height       int
	Bandwidth    int
	PlaylistPath string
}

func (p *EncodingPipeline) Encode(ctx context.Context, s3c *storage.S3Client) error {
	log.Printf("[%s] Stage [3/5]: Encoding...\n", p.Payload.VideoID)

	var renditionsToEncode []int
	for _, res := range p.Payload.Resolutions {
		if res <= p.SourceInfo.Height {
			renditionsToEncode = append(renditionsToEncode, res)
		}
	}

	sourceResInList := slices.Contains(renditionsToEncode, p.SourceInfo.Height)

	if !sourceResInList {
		renditionsToEncode = append(renditionsToEncode, p.SourceInfo.Height)
	}

	sort.Ints(renditionsToEncode)

	if len(renditionsToEncode) == 0 {
		return fmt.Errorf("no renditions to produce for source height %d", p.SourceInfo.Height)
	}

	var wg sync.WaitGroup // this is for encoding goroutines
	var mu sync.Mutex     // this is for master playlist updating mutex

	// limit the number of concurrent transcoding
	maxConcurrent := 2
	sem := make(chan struct{}, maxConcurrent)

	var completedRenditions []completedRendition
	var encodingErrors []error
	var firstRenditionNotified bool

	hlsBase := filepath.Join(p.EncodedOutputPath, "hls")
	if err := os.MkdirAll(hlsBase, 0o755); err != nil {
		return fmt.Errorf("failed to create hls base dir: %w", err)
	}

	log.Printf("[%s] Starting encoding for renditions: %v\n", p.Payload.VideoID, renditionsToEncode)

	for _, height := range renditionsToEncode {
		wg.Add(1)

		go func(height int) {
			sem <- struct{}{}
			defer func() { <-sem }()
			defer wg.Done()

			err := p.EncodeRendition(ctx, height)

			if err != nil {
				log.Printf("[%s] ERROR encoding %dp: %v\n", p.Payload.VideoID, height, err)
				mu.Lock()
				encodingErrors = append(encodingErrors, fmt.Errorf("failed on %dp: %w", height, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			defer mu.Unlock()

			scaledWidth := (p.SourceInfo.Width * height) / p.SourceInfo.Height
			if scaledWidth%2 != 0 {
				scaledWidth++
			}

			completedRenditions = append(completedRenditions, completedRendition{
				Height:       height,
				Width:        scaledWidth,
				Bandwidth:    getBandwidthForHeight(height),
				PlaylistPath: fmt.Sprintf("%dp/playlist.m3u8", height),
			})

			if err := p.updateMasterPlaylist(ctx, hlsBase, completedRenditions); err != nil {
				log.Printf("[%s] ERROR updating master playlist after %dp rendition: %v\n", p.Payload.VideoID, height, err)
				encodingErrors = append(encodingErrors, fmt.Errorf("failed to update master playlist for %dp: %w", height, err))
			}

			if !firstRenditionNotified && p.Uploader != nil {
				firstRenditionNotified = true
				p.Uploader.NotifyPlayable()
			}
		}(height)

	}

	wg.Wait()

	if len(encodingErrors) > 0 {
		return fmt.Errorf("encountered %d error(s) during encoding: %v", len(encodingErrors), encodingErrors)
	}

	log.Printf("[%s] Stage [3/5]: All encoding tasks finished.\n", p.Payload.VideoID)
	return nil

}

func (p *EncodingPipeline) Upload(ctx context.Context) error {
	log.Printf("[%s] Stage [4/5]: Uploading to S3...\n", p.Payload.VideoID)

	return filepath.Walk(p.EncodedOutputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Do not upload master playlist here, as this will be automatically uploaded on updateMasterPlaylist
		// which is called by the rendition threads
		if !info.IsDir() && filepath.Base(path) != "master.m3u8" {
			relativePath, err := filepath.Rel(p.EncodedOutputPath, path)
			if err != nil {
				return err
			}

			objectKey := filepath.Join(p.Payload.VideoID, relativePath)

			log.Printf("Uploading %s to %s", path, objectKey)
			if err := p.Uploader.Upload(ctx, path, objectKey); err != nil {
				return fmt.Errorf("failed to upload %s: %w", info.Name(), err)
			}
		}
		return nil
	})
}

func (p *EncodingPipeline) EncodeRendition(ctx context.Context, height int) error {
	renditionDir := filepath.Join(p.EncodedOutputPath, "hls", fmt.Sprintf("%dp", height))

	if err := os.MkdirAll(renditionDir, 0o755); err != nil {
		log.Printf("Failed to create rendition directory %s: %v", renditionDir, err)
		return fmt.Errorf("failed to create rendition directory %s: %w", renditionDir, err)
	}

	// This is hacky, but we need some way to define the bitrate
	audioBitrate := chooseAudioBitrate(height)

	args := []string{
		"-hide_banner", "-y",
		"-i", p.StreamURL,
		"-c:v", getBestVideoEncoder(),
		"-profile:v", "main",
		"-pix_fmt", "yuv420p",
		"-vf", fmt.Sprintf("scale=-2:%d", height),
	}

	args = append(args, getQualityArgs()...)

	if p.SourceInfo.HasAudio {
		args = append(args,
			"-c:a", "aac",
			"-b:a", audioBitrate,
			"-map", "0:v:0", // Video
			"-map", "0:a:0", // Audio
		)
	}

	args = append(args,
		"-f", "hls",
		"-hls_time", "4", // HLS TIME CHUNK DURATION
		"-hls_playlist_type", "vod",
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(renditionDir, "segment%03d.ts"),
		filepath.Join(renditionDir, "playlist.m3u8"),
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	log.Printf("[%s] Encoding %dp: ffmpeg %s\n", p.Payload.VideoID, height, strings.Join(args, " "))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed for %dp: %w\n--- FFmpeg output ---\n%s", height, err, stderr.String())
	}

	log.Printf("[%s] Finished encoding %dp\n", p.Payload.VideoID, height)
	return nil

}

func (p *EncodingPipeline) updateMasterPlaylist(ctx context.Context, hlsBaseDir string, renditions []completedRendition) error {
	masterPlaylistPath := filepath.Join(hlsBaseDir, "master.m3u8")

	log.Printf("[%s] Updating master playlist at %s\n", p.Payload.VideoID, masterPlaylistPath)

	sort.Slice(renditions, func(i, j int) bool {
		return renditions[i].Height < renditions[j].Height
	})

	var content strings.Builder
	content.WriteString("#EXTM3U\n")
	content.WriteString("#EXT-X-VERSION:3\n")

	for _, r := range renditions {
		content.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n", r.Bandwidth, r.Width, r.Height))
		content.WriteString(r.PlaylistPath + "\n")
	}

	if err := os.WriteFile(masterPlaylistPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write master playlist: %w", err)
	}

	objectKey := filepath.Join(p.Payload.VideoID, "hls", "master.m3u8")
	if err := p.Uploader.Upload(ctx, masterPlaylistPath, objectKey); err != nil {
		return fmt.Errorf("failed to upload master playlist: %w", err)
	}

	log.Printf("[%s] Successfully updated and uploaded master playlist with %d rendition(s).\n", p.Payload.VideoID, len(renditions))
	return nil

}

func (p *EncodingPipeline) Cleanup() error {
	log.Println("Stage: Cleanup...")
	if err := os.RemoveAll(p.TempDir); err != nil {
		return fmt.Errorf("failed to remove temp dir %s: %w", p.TempDir, err)
	}
	return nil
}
