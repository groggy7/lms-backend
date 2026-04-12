package repository

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/serhatkilbas/lms-poc/internal/domain"
)

type ffmpegTranscoder struct{}

func NewFFmpegTranscoder() domain.Transcoder {
	return &ffmpegTranscoder{}
}

func (t *ffmpegTranscoder) TranscodeToHLS(ctx context.Context, inputPath, outputDir string) (string, error) {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return "", err
		}
	}

	playlistPath := filepath.Join(outputDir, "index.m3u8")

	// FFmpeg command for HLS generation
	// -i inputPath: Input file
	// -profile:v baseline: Compatibility profile
	// -level 3.0: Level 3.0 for better compatibility
	// -s 1280x720: Resize to 720p (optional)
	// -start_number 0: Segment start index
	// -hls_time 10: Segment duration in seconds
	// -hls_list_size 0: Include all segments in the playlist
	// -f hls: Output format HLS
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", inputPath,
		"-profile:v", "baseline",
		"-level", "3.0",
		"-start_number", "0",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-f", "hls",
		playlistPath,
	)

	// Combine stdout and stderr for better debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg error: %v, output: %s", err, string(output))
	}

	return playlistPath, nil
}
