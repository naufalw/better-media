package livestream

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/nareix/joy5/format/flv"
	"github.com/nareix/joy5/format/rtmp"
)

// represent an active livestream
type Stream struct {
	Key           string
	StartedAt     time.Time
	OutputDir     string
	HLSPath       string
	RecordingPath string
	cancel        context.CancelFunc
}

// handle rtmp ingest
type RTMPServer struct {
	Port          int
	OutputBase    string
	streams       map[string]*Stream
	streamsMu     sync.RWMutex
	OnStreamStart func(streamKey string, hlsPath string)
	OnStreamEnd   func(streamKey string)
	ValidateKey   func(key string) bool

	OnRecordReady func(streamKey string, outputDir string)
}

func NewRTMPServer(port int, outputBase string) *RTMPServer {
	return &RTMPServer{
		Port:       port,
		OutputBase: outputBase,
		streams:    make(map[string]*Stream),
	}
}

func (s *RTMPServer) Start() error {

	// open TCP
	addr := fmt.Sprintf(":%d", s.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	log.Printf("[RTMP] Server listening on rtmp://localhost%s", addr)
	log.Printf("[RTMP] OBS Settings: Server=rtmp://localhost%s/live  Stream Key=<your-key>", addr)

	// make rtmp
	server := rtmp.NewServer()
	server.LogEvent = func(c *rtmp.Conn, nc net.Conn, e int) {
		log.Printf("[RTMP] Event: %s", rtmp.EventString[e])
	}
	server.HandleConn = func(c *rtmp.Conn, nc net.Conn) {
		s.handleStream(c, nc)
	}

	go func() {
		for {
			// tcp comes, accept, handover to the rtmp
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("[RTMP] Accept error: %v", err)
				continue
			}
			go server.HandleNetConn(conn)
		}
	}()
	return nil

}

// RTMP packets → FLV Muxer → pipe → FFmpeg stdin → HLS output
// C is rtmp, NC is the TCP socket
func (s *RTMPServer) handleStream(c *rtmp.Conn, nc net.Conn) {
	defer nc.Close()

	// streamkey from URL (OBS: rtmp://server/live/stream-key)
	streamKey := filepath.Base(c.URL.Path)
	if streamKey == "" || streamKey == "live" || streamKey == "/" {
		streamKey = fmt.Sprintf("stream_%d", time.Now().Unix())
	}

	log.Printf("[RTMP] Stream starting: %s (publishing=%v)", streamKey, c.Publishing)
	if s.ValidateKey != nil && !s.ValidateKey(streamKey) {
		log.Printf("[RTMP] Invalid stream key: %s", streamKey)
		return
	}

	if !c.Publishing {
		log.Printf("[RTMP] Not a publish stream, ignoring")
		return
	}

	outputDir := filepath.Join(s.OutputBase, streamKey)

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Printf("[RTMP] Failed to create output dir: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	hlsPath := filepath.Join(outputDir, "index.m3u8")
	recordingPath := filepath.Join(outputDir, "recording.mkv")

	// register the stream to hashmap
	stream := &Stream{
		Key:           streamKey,
		StartedAt:     time.Now(),
		OutputDir:     outputDir,
		HLSPath:       hlsPath,
		RecordingPath: recordingPath,
		cancel:        cancel,
	}
	s.streamsMu.Lock()
	s.streams[streamKey] = stream
	s.streamsMu.Unlock()

	defer func() {
		s.streamsMu.Lock()
		delete(s.streams, streamKey)
		s.streamsMu.Unlock()
		if s.OnStreamEnd != nil {
			s.OnStreamEnd(streamKey)
		}
		cancel()
	}()

	// Transcoding
	segmentPath := filepath.Join(outputDir, "segment%03d.ts")
	ffmpegArgs := []string{
		"-hide_banner",
		"-i", "pipe:0", // FLV is in stdin

		// OUTPUT 1 => HLS
		"-c:v", "copy", // OBS already encoded it, so just need to copy
		"-c:a", "aac",
		"-f", "hls",
		"-hls_time", "2", // segment 2 second for low latency
		"-hls_list_size", "10",
		"-hls_flags", "delete_segments+append_list",
		"-hls_segment_filename", segmentPath,
		hlsPath,

		// OUTPUT 2 => MKV
		"-c:v", "copy",
		"-c:a", "aac",
		recordingPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", ffmpegArgs...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("[RTMP] Failed to get stdin pipe: %v", err)
		return
	}

	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Printf("[RTMP] Failed to start FFmpeg: %v", err)
		return
	}

	log.Printf("[RTMP] Stream %s started, HLS: %s", streamKey, hlsPath)
	if s.OnStreamStart != nil {
		s.OnStreamStart(streamKey, hlsPath)
	}

	// RTMP packet is basically FLV tag, here we remux the rtmp to flv so ffmpeg can get it natively
	// ffmpeg will get the remuxed FLV via stdin
	muxer := flv.NewMuxer(stdin)
	muxer.WriteFileHeader()

	// Read RTMP packets and write to FLV
	for {
		pkt, err := c.ReadPacket()
		if err != nil {
			if err != io.EOF {
				log.Printf("[RTMP] Read error: %v", err)
			}
			break
		}
		if err := muxer.WritePacket(pkt); err != nil {
			log.Printf("[RTMP] Write error: %v", err)
			break
		}
	}
	stdin.Close()
	cmd.Wait()

	// trigger recording
	if s.OnRecordReady != nil {
		s.OnRecordReady(streamKey, outputDir)
	}
	log.Printf("[RTMP] Stream %s ended", streamKey)

}

// Returns all active stream keys
func (s *RTMPServer) GetActiveStreams() []string {
	s.streamsMu.RLock()
	defer s.streamsMu.RUnlock()
	keys := make([]string, 0, len(s.streams))
	for k := range s.streams {
		keys = append(keys, k)
	}
	return keys
}

// Get a stream
func (s *RTMPServer) GetStream(key string) *Stream {
	s.streamsMu.RLock()
	defer s.streamsMu.RUnlock()
	return s.streams[key]
}
