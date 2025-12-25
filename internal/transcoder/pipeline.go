package transcoder

import (
	"better-media/internal/storage"
	"better-media/internal/transcription"
	"better-media/internal/uploader"
	"better-media/pkg/models"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
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
		Duration float64
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

	if p.StreamURL == "" {
		if s3c == nil {
			return fmt.Errorf("StreamURL is empty and no S3 client provided")
		}
		objectKey := filepath.Join(p.Payload.VideoID, "source", p.Payload.InputFile)
		presignedGet, err := s3c.GeneratePresignedGet(ctx, objectKey, time.Hour)
		if err != nil {
			return fmt.Errorf("failed to generate presigned URL: %w", err)
		}
		p.StreamURL = presignedGet.URL
	}

	defer p.Cleanup()

	if err := p.Probe(); err != nil {
		return fmt.Errorf("failed to probe file: %w", err)
	}

	if err := p.GenerateThumbnails(ctx); err != nil {
		log.Printf("[%s] Warning: Thumbnail generation failed: %v", p.Payload.VideoID, err)
		// if thumbnail fail, just ignore
	}

	if err := p.GenerateSubtitles(ctx); err != nil {
		log.Printf("[%s] Warning: Subtitle generation failed: %v", p.Payload.VideoID, err)
		// if subtitle fail, just ignore
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

	// getting the duration
	if formatMap, ok := data["format"].(map[string]any); ok {
		if durationStr, ok := formatMap["duration"].(string); ok {
			if duration, err := strconv.ParseFloat(durationStr, 64); err == nil {
				p.SourceInfo.Duration = duration
			}
		}
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

// Encode, upload, and update the master playlist for a single rendition
func (p *EncodingPipeline) encodeAndUploadRendition(
	ctx context.Context,
	height int,
	completedRenditions *[]completedRendition,
	mu *sync.Mutex,
	hlsBase string,
) error {
	err := p.EncodeRendition(ctx, height)
	if err != nil {
		return fmt.Errorf("encoding %dp failed: %w", height, err)
	}
	log.Printf("[%s] Finished encoding %dp\n", p.Payload.VideoID, height)

	renditionDir := filepath.Join(p.EncodedOutputPath, "hls", fmt.Sprintf("%dp", height))
	if err := p.uploadRendition(ctx, renditionDir, height); err != nil {
		return fmt.Errorf("uploading %dp failed: %w", height, err)
	}

	mu.Lock()
	defer mu.Unlock()

	scaledWidth := (p.SourceInfo.Width * height) / p.SourceInfo.Height
	if scaledWidth%2 != 0 {
		scaledWidth++
	}

	bw := calculateRenditionBandwidth(renditionDir, p.SourceInfo.Duration)

	*completedRenditions = append(*completedRenditions, completedRendition{
		Height:       height,
		Width:        scaledWidth,
		Bandwidth:    bw,
		PlaylistPath: fmt.Sprintf("%dp/playlist.m3u8", height),
	})

	if err := p.updateMasterPlaylist(ctx, hlsBase, *completedRenditions); err != nil {
		return fmt.Errorf("updating master playlist failed: %w", err)
	}

	return nil
}

func (p *EncodingPipeline) EncodeMultipleRenditions(ctx context.Context, heights []int) error {
	if len(heights) == 0 {
		return nil
	}

	encoder := getBestVideoEncoder()
	qualityArgs := getQualityArgs()

	// we are building filter complex
	var filterParts []string
	splitCount := len(heights)

	// split filter: [0:v]split=N[v1][v2][v3]...
	splitOutputs := ""
	for i := range heights {
		splitOutputs += fmt.Sprintf("[v%d]", i)
	}
	filterParts = append(filterParts, fmt.Sprintf("[0:v]split=%d%s", splitCount, splitOutputs))

	// scale filter : [v0]scale=-2:360[360p]; [v1]scale=-2:720[720p]...
	for i, h := range heights {
		filterParts = append(filterParts, fmt.Sprintf("[v%d]scale=-2:%d[%dp]", i, h, h))
	}

	// join filters
	filterComplex := strings.Join(filterParts, "; ")
	args := []string{
		"-hide_banner", "-y",
		"-i", p.StreamURL,
		"-filter_complex", filterComplex,
	}

	// process filepaths
	for _, h := range heights {
		renditionDir := filepath.Join(p.EncodedOutputPath, "hls", fmt.Sprintf("%dp", h))
		if err := os.MkdirAll(renditionDir, 0o755); err != nil {
			return fmt.Errorf("failed to create rendition directory %s: %w", renditionDir, err)
		}

		playlistPath := filepath.Join(renditionDir, "playlist.m3u8")
		segmentPattern := filepath.Join(renditionDir, "segment%03d.ts")

		args = append(args,
			"-map", fmt.Sprintf("[%dp]", h),
			"-c:v", encoder,
			"-profile:v", "main",
			"-pix_fmt", "yuv420p",
		)
		args = append(args, qualityArgs...)
		args = append(args,
			"-c:a", "aac",
			"-b:a", chooseAudioBitrate(h),
			"-f", "hls",
			"-hls_time", "4",
			"-hls_playlist_type", "vod",
			"-hls_list_size", "0",
			"-hls_segment_filename", segmentPattern,
			playlistPath,
		)
	}

	if p.SourceInfo.HasAudio {
		args = append(args, "-map", "0:a:0")
	}

	log.Printf("[%s] Transcoding for rest renditions %v: ffmpeg %s", p.Payload.VideoID, heights, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Parse progress from stderr
	go func() {
		parseFFmpegProgress(stderr, func(currentSeconds float64) {
			if p.SourceInfo.Duration > 0 && p.Uploader != nil {
				// multioutput start after first rendition
				totalRenditions := len(selectResolutions(p.SourceInfo.Height))
				baseProgress := 100 / totalRenditions // first rendition
				remainingShare := 100 - baseProgress  // What's left for multiout

				// scale current progress to remaining portion
				phasePercent := int((currentSeconds / p.SourceInfo.Duration) * float64(remainingShare))
				percent := baseProgress + phasePercent
				if percent > 100 {
					percent = 100
				}
				p.Uploader.UpdateProgress(percent)
			}
		})
	}()
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("multi encoding failed: %w", err)
	}

	return nil

}

// Get the thumbnail of the video by extracting image at 10% of the duration
func (p *EncodingPipeline) GenerateThumbnails(ctx context.Context) error {
	timestamp := p.SourceInfo.Duration * 0.1
	if timestamp < 1 {
		timestamp = 1
	}

	sizes := []int{320, 640, 1280}
	thumbDir := filepath.Join(p.EncodedOutputPath, "thumbnails")
	if err := os.MkdirAll(thumbDir, 0o755); err != nil {
		return fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	for _, size := range sizes {
		outPath := filepath.Join(thumbDir, fmt.Sprintf("thumb_%d.jpg", size))
		args := []string{
			"-ss", fmt.Sprintf("%.2f", timestamp),
			"-i", p.StreamURL,
			"-vframes", "1",
			"-vf", fmt.Sprintf("scale=%d:-1", size),
			"-q:v", "2",
			outPath,
		}

		cmd := exec.CommandContext(ctx, "ffmpeg", args...)
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to generate %dpx thumbnail: %v", size, err)
			continue
		}
	}

	return nil
}

// Generate subtitle
func (p *EncodingPipeline) GenerateSubtitles(ctx context.Context) error {
	if !p.Payload.Transcribe {
		log.Printf("[%s] Skipping subtitles: transcription not requested", p.Payload.VideoID)
		return nil
	}

	apiURL := os.Getenv("TRANSCRIPTION_API_URL")
	apiKey := os.Getenv("TRANSCRIPTION_API_KEY")
	model := os.Getenv("TRANSCRIPTION_MODEL")
	if apiURL == "" || apiKey == "" {
		log.Printf("[%s] Skipping subtitles: TRANSCRIPTION_API_URL or TRANSCRIPTION_API_KEY not set", p.Payload.VideoID)
		return nil
	}
	if model == "" {
		model = "whisper-large-v3"
	}
	log.Printf("[%s] Generating subtitles via %s", p.Payload.VideoID, apiURL)
	provider := transcription.NewProvider(apiURL, apiKey, model)
	transcriber := transcription.NewTranscriber(provider)
	subtitlesDir := filepath.Join(p.EncodedOutputPath, "subtitles")
	vttPath, err := transcriber.TranscribeVideo(ctx, p.StreamURL, subtitlesDir)
	if err != nil {
		return fmt.Errorf("subtitle generation failed: %w", err)
	}
	log.Printf("[%s] Subtitles generated: %s", p.Payload.VideoID, vttPath)
	return nil
}

func (p *EncodingPipeline) Encode(ctx context.Context, s3c *storage.S3Client) error {
	log.Printf("[%s] Stage [3/5]: Encoding...\n", p.Payload.VideoID)
	renditions := selectResolutions(p.SourceInfo.Height)
	log.Printf("[%s] Starting encoding for renditions: %v\n", p.Payload.VideoID, renditions)

	var mu sync.Mutex
	var completedRenditions []completedRendition
	var encodingErrors []error

	hlsBase := filepath.Join(p.EncodedOutputPath, "hls")
	if err := os.MkdirAll(hlsBase, 0o755); err != nil {
		return fmt.Errorf("failed to create hls directory: %w", err)
	}

	sort.Ints(renditions)

	//Encode lowest first for early playback
	if len(renditions) > 0 {
		firstHeight := renditions[0]
		log.Printf("[%s] Encoding %dp first for early playback", p.Payload.VideoID, firstHeight)
		if err := p.encodeAndUploadRendition(ctx, firstHeight, &completedRenditions, &mu, hlsBase); err != nil {
			encodingErrors = append(encodingErrors, err)
		} else {
			// Notify playable after first rendition
			if p.Uploader != nil {
				p.Uploader.NotifyPlayable()
			}
		}
	}
	// Encode remaining renditions using single ffmpeg with multiple outputs

	if len(renditions) > 1 {
		remaining := renditions[1:]

		log.Printf("[%s] single ffmpeg multi out encoding for: %v", p.Payload.VideoID, remaining)

		if err := p.EncodeMultipleRenditions(ctx, remaining); err != nil {
			encodingErrors = append(encodingErrors, err)
		} else {
			for _, height := range remaining {
				renditionDir := filepath.Join(p.EncodedOutputPath, "hls", fmt.Sprintf("%dp", height))

				if err := p.uploadRendition(ctx, renditionDir, height); err != nil {
					encodingErrors = append(encodingErrors, err)
					continue
				}

				scaledWidth := (p.SourceInfo.Width * height) / p.SourceInfo.Height
				if scaledWidth%2 != 0 {
					scaledWidth++
				}
				bw := calculateRenditionBandwidth(renditionDir, p.SourceInfo.Duration)

				mu.Lock()
				completedRenditions = append(completedRenditions, completedRendition{
					Height:       height,
					Width:        scaledWidth,
					Bandwidth:    bw,
					PlaylistPath: fmt.Sprintf("%dp/playlist.m3u8", height),
				})
				if err := p.updateMasterPlaylist(ctx, hlsBase, completedRenditions); err != nil {
					encodingErrors = append(encodingErrors, err)
				}
				mu.Unlock()
			}
		}
	}

	log.Printf("[%s] Stage [3/5]: All encoding tasks finished.\n", p.Payload.VideoID)
	if len(encodingErrors) > 0 {
		return fmt.Errorf("encountered %d error(s) during encoding: %v", len(encodingErrors), encodingErrors)
	}
	return nil
}

func (p *EncodingPipeline) uploadRendition(ctx context.Context, renditionDir string, height int) error {
	return filepath.Walk(renditionDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		relativePath, _ := filepath.Rel(p.EncodedOutputPath, path)
		objectKey := filepath.Join(p.Payload.VideoID, relativePath)

		log.Printf("Uploading %s to %s", path, objectKey)
		return p.Uploader.Upload(ctx, path, objectKey)
	})
}

func (p *EncodingPipeline) Upload(ctx context.Context) error {
	log.Printf("[%s] Stage [4/5]: Uploading to S3...\n", p.Payload.VideoID)

	// thumbnail
	thumbDir := filepath.Join(p.EncodedOutputPath, "thumbnails")
	if _, err := os.Stat(thumbDir); err == nil {
		if err := filepath.Walk(thumbDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			objectKey := filepath.Join(p.Payload.VideoID, "thumbnails", info.Name())
			log.Printf("Uploading thumbnail %s", objectKey)
			if err := p.Uploader.Upload(ctx, path, objectKey); err != nil {
				log.Printf("Failed to upload thumbnail %s: %v", objectKey, err)
				// Continue with other thumbnails
			}
			return nil
		}); err != nil {
			log.Printf("Error walking thumbnail directory: %v", err)
		}
	}

	//subtitle
	subtitlesDir := filepath.Join(p.EncodedOutputPath, "subtitles")
	if _, err := os.Stat(subtitlesDir); err == nil {
		if err := filepath.Walk(subtitlesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			objectKey := filepath.Join(p.Payload.VideoID, "subtitles", info.Name())
			log.Printf("Uploading subtitle %s", objectKey)
			if err := p.Uploader.Upload(ctx, path, objectKey); err != nil {
				log.Printf("Failed to upload subtitle %s: %v", objectKey, err)
			}
			return nil
		}); err != nil {
			log.Printf("Error walking subtitles directory: %v", err)
		}
	}

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

	log.Printf("[%s] Encoding %dp: ffmpeg %s\n", p.Payload.VideoID, height, strings.Join(args, " "))

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// get the progress from piped stderr
	go func() {
		parseFFmpegProgress(stderr, func(currentSeconds float64) {
			if p.SourceInfo.Duration > 0 && p.Uploader != nil {
				totalRenditions := len(selectResolutions(p.SourceInfo.Height))
				perRenditionPercent := 100 / totalRenditions
				percent := int((currentSeconds / p.SourceInfo.Duration) * float64(perRenditionPercent))
				if percent > 100 {
					percent = 100
				}
				p.Uploader.UpdateProgress(percent)
			}
		})
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg failed for %dp: %w", height, err)
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
