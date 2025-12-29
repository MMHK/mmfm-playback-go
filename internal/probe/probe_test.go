package probe

import (
	"mmfm-playback-go/internal/config"
	"mmfm-playback-go/tests"
	"os"
	"testing"
)

var conf *config.PlaybackConfig

func init() {
	tests.LoadTestEnv()

	conf, _ = config.NewConfig(tests.GetLocalPath("../config.json"))
}

func TestNewFFprobe(t *testing.T) {
	if conf == nil {
		t.Fatal("Expected PlaybackConfig instance, got nil")
	}

	ffprobe := NewFFprobe(conf.FFMpegConf.FFProbe)
	if ffprobe == nil {
		t.Fatal("Expected FFprobe instance, got nil")
	}

	envFFProbe := os.Getenv("FFPROBE_PATH")

	if ffprobe.bin != envFFProbe {
		t.Errorf("Expected bin to be '/usr/bin/ffprobe', got '%s'", ffprobe.bin)
	}
}

func TestMediaInfoGetDuration(t *testing.T) {
	ffprobe := NewFFprobe(conf.FFMpegConf.FFProbe)
	if ffprobe == nil {
		t.Fatal("Expected FFprobe instance, got nil")
	}

	testENVAsset := os.Getenv("TEST_ASSET")
	if testENVAsset == "" {
		t.Skip("Skipping test, TEST_ASSET environment variable is not set")
	} else {
		t.Logf("Testing with asset: %s", testENVAsset)
	}

	// This is a placeholder test since the GetDuration method currently returns 0, nil
	mediaInfo, err := ffprobe.GetMediaInfo(testENVAsset)
	if err != nil {
		t.Errorf("GetMediaInfo should not return error, got: %v", err)
	} else {
		t.Logf("MediaInfo: %s", tests.ToJSON(mediaInfo))
	}

	duration, err := mediaInfo.GetDuration()
	if err != nil {
		t.Errorf("GetDuration should not return error, got: %v", err)
	} else {
		t.Logf("Duration: %f", duration)
	}

	if duration <= 0 {
		t.Errorf("Expected duration to be greater than 0, got %f", duration)
	}
}
