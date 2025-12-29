package probe

import (
	"fmt"
	"mmfm-playback-go/internal/logger"
	"os/exec"
	"strconv"
	"strings"
)

// FFprobe represents the ffprobe wrapper
type FFprobe struct {
	bin string
}

// NewFFprobe creates a new FFprobe instance
func NewFFprobe(bin string) *FFprobe {
	return &FFprobe{
		bin: bin,
	}
}

// GetMediaInfo retrieves media information
func (f *FFprobe) GetMediaInfo(url string) (*MediaInfo, error) {
	cmd := exec.Command(f.bin, "-v", "quiet", "-show_format", "-show_streams", url)
	output, err := cmd.Output()
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// Parse the output to extract media information
	// This is a simplified implementation
	return &MediaInfo{raw: string(output)}, nil
}

// MediaInfo holds media information
type MediaInfo struct {
	raw string
}

// GetDuration retrieves the duration of the media file
func (mi *MediaInfo) GetDuration() (float64, error) {
	// Parse the raw output to extract the duration
	lines := strings.Split(mi.raw, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "duration=") {
			// Extract the duration value from FORMAT section
			durationStr := strings.TrimSpace(strings.TrimPrefix(line, "duration="))
			duration, err := strconv.ParseFloat(durationStr, 64)
			if err != nil {
				return 0, err
			}
			return duration, nil
		}
	}

	return 0, fmt.Errorf("duration not found in output")
}
