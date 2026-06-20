package player

import (
	"fmt"
	"log/slog"
	"os/exec"
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
	slog.Debug("mplayer Play", "bin", m.bin, "url", url, "second", second)
	if m.cmd != nil {
		slog.Debug("mplayer stopping previous instance")
		m.Stop()
	}

	args := []string{}
	if second > 0 {
		args = append(args, "-ss", SecToString(second), "-vo", "null", "-slave", fmt.Sprintf("%ds", second))
	}
	args = append(args, url)
	slog.Debug("mplayer Play args", "args", args)

	m.cmd = exec.Command(m.bin, args...)
	err := m.cmd.Start()
	if err != nil {
		slog.Error("mplayer Play start failed", "error", err)
		return nil, err
	}
	slog.Debug("mplayer Play started", "pid", m.cmd.Process.Pid)

	done := make(chan bool, 1)
	go func() {
		m.cmd.Wait()
		done <- true
	}()

	return done, nil
}

// Stop stops the current playback
func (m *Mplayer) Stop() error {
	slog.Debug("mplayer Stop")
	if m.cmd != nil && m.cmd.Process != nil {
		slog.Debug("mplayer killing process", "pid", m.cmd.Process.Pid)
		if err := m.cmd.Process.Kill(); err != nil {
			slog.Error("mplayer kill failed", "error", err)
		}
	}
	m.cmd = nil
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
	slog.Debug("ffprobe GetMediaInfo", "bin", f.bin, "url", url)
	cmd := exec.Command(f.bin, "-v", "quiet", "-show_format", "-show_streams", url)
	output, err := cmd.Output()
	if err != nil {
		slog.Error("ffprobe GetMediaInfo failed", "error", err)
		return nil, err
	}
	slog.Debug("ffprobe GetMediaInfo succeeded", "output_size", len(output))

	return &MediaInfo{raw: string(output)}, nil
}

// MediaInfo holds media information
type MediaInfo struct {
	raw string
}

// GetDuration retrieves the duration of the media file
func (mi *MediaInfo) GetDuration() (float64, error) {
	lines := strings.Split(mi.raw, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "duration=") {
			durationStr := strings.TrimSpace(strings.TrimPrefix(line, "duration="))
			duration, err := strconv.ParseFloat(durationStr, 64)
			if err != nil {
				slog.Error("GetDuration parse failed", "value", durationStr, "error", err)
				return 0, err
			}
			slog.Debug("GetDuration succeeded", "duration", duration)
			return duration, nil
		}
	}
	slog.Debug("GetDuration no duration found in media info")
	return 0, nil
}
