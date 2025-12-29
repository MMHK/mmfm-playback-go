package player

import (
	"mmfm-playback-go/internal/config"
	"mmfm-playback-go/pkg/types"
	"mmfm-playback-go/tests"
	"os"
	"testing"
)

var conf *config.PlaybackConfig

func init() {
	tests.LoadTestEnv()
	conf, _ = config.NewConfig(tests.GetLocalPath("../config.json"))
}

func TestNewMusicPlayer(t *testing.T) {
	conf := &config.PlaybackConfig{
		FFMpegConf: &config.FFmpegConfig{
			FFPlay:  "/usr/bin/ffplay",
			FFProbe: "/usr/bin/ffprobe",
			MPlayer: "/usr/bin/mplayer",
		},
		WebSocketAPI: "ws://localhost:8888",
		WebAPI:       "http://localhost:8888/song/get",
		CachePath:    "./cache",
	}

	player := NewMusicPlayer(conf)
	if player == nil {
		t.Fatal("Expected MusicPlayer instance, got nil")
	}

	if player.Conf != conf {
		t.Error("MusicPlayer should hold the provided config")
	}
}

func TestMusicPlayer_Play(t *testing.T) {
	if conf == nil {
		t.Skip("Skipping test because conf is nil")
	}
	player := NewMusicPlayer(conf)

	testENVAsset := os.Getenv("TEST_ASSET")
	if testENVAsset == "" {
		t.Skip("Skipping test, TEST_ASSET environment variable is not set")
	} else {
		t.Logf("Testing with asset: %s", testENVAsset)
	}

	err := player.Play(&types.Song{
		URL:    testENVAsset,
		Name:   "test",
		Author: "test",
	}, 0)

	if err != nil {
		t.Error("Expected no error, got:", err)
	}

}

func TestMusicPlayer_Start(t *testing.T) {
	if conf == nil {
		t.Skip("Skipping test because conf is nil")
	}
	player := NewMusicPlayer(conf)

	err := player.Start()
	if err != nil {
		t.Error("Expected no error, got:", err)
	}
}
