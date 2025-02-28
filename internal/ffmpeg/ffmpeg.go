package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
)

// RunFFmpeg executes an FFmpeg command with the provided arguments
func (s *service) RunFFmpeg(ctx context.Context, args []string) ([]byte, error) {
	if err := s.EnsureInstalled(); err != nil {
		return nil, fmt.Errorf("failed to ensure FFmpeg is installed: %w", err)
	}

	cmd := exec.CommandContext(ctx, s.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return output, fmt.Errorf("ffmpeg command failed: %w", err)
	}

	return output, nil
}
