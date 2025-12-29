package player

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Mplayer represents the mplayer wrapper
type Mplayer struct {
	bin string
	cmd *exec.Cmd
}

// NewMplayer creates a new Mplayer instance
func NewMplayer(bin string) *Mplayer {
	return &Mplayer{
		bin: bin,
	}
}

func SecToString(second int) string {
	// hh:mm:ss
	return fmt.Sprintf("%02d:%02d:%02d", second/3600, (second/60)%60, second%60)
}

// Play plays a media file from a specific time
func (m *Mplayer) Play(url string, second int) (<-chan bool, error) {
	if m.cmd != nil {
		m.Stop()
	}

	args := []string{}
	if second > 0 {
		args = append(args, "-ss", SecToString(second), "-vo", "null", "-slave", fmt.Sprintf("%ds", second))
	}
	args = append(args, url)

	m.cmd = exec.Command(m.bin, args...)
	err := m.cmd.Start()
	if err != nil {
		return nil, err
	}

	done := make(chan bool, 1)
	go func() {
		m.cmd.Wait()
		done <- true
	}()

	return done, nil
}

// Stop stops the current playback
func (m *Mplayer) Stop() error {
	if m.cmd != nil {
		if runtime.GOOS == "windows" {
			m.cmd.Process.Kill()
		} else {
			m.cmd.Process.Kill()
		}
		m.cmd = nil
	}
	return nil
}

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
	// Simplified implementation - in a real scenario, you would parse the raw output
	// to extract the duration field
	lines := strings.Split(mi.raw, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "duration=") {
			durationStr := strings.TrimSpace(strings.TrimPrefix(line, "duration="))
			duration, err := strconv.ParseFloat(durationStr, 64)
			if err != nil {
				return 0, err
			}
			return duration, nil
		}
	}
	return 0, nil
}
