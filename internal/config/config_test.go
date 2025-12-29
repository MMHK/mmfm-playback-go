package config

import (
	"mmfm-playback-go/tests"
	"os"
	"testing"
)

func TestNewConfigFromFile(t *testing.T) {
	// Create a temporary config file
	tempConfig := `{
    "ffmpeg": {
        "ffplay": "/usr/bin/ffplay",
        "ffprobe": "/usr/bin/ffprobe",
        "mplayer": "/usr/bin/mplayer"
    },
    "ws": "ws://localhost:8888",
    "web": "http://localhost:8888/song/get",
    "cache": "./cache"
}`

	tempFile := "test_config.json"
	err := os.WriteFile(tempFile, []byte(tempConfig), 0644)
	if err != nil {
		t.Fatal("Failed to create temp config file:", err)
	}
	defer os.Remove(tempFile) // clean up

	config, err := NewConfig(tempFile)
	if err != nil {
		t.Fatal("Failed to load config:", err)
	}

	if config.FFMpegConf.FFPlay != "/usr/bin/ffplay" {
		t.Errorf("Expected ffplay path to be '/usr/bin/ffplay', got '%s'", config.FFMpegConf.FFPlay)
	}

	if config.WebSocketAPI != "ws://localhost:8888" {
		t.Errorf("Expected WebSocket API to be 'ws://localhost:8888', got '%s'", config.WebSocketAPI)
	}
}

func TestConfigFromEnvironment(t *testing.T) {
	err := tests.LoadTestEnv()
	if err != nil {
		t.Fatal("Failed to load test env:", err)
	}

	// Create a minimal config file
	tempConfig := `{
    "ffmpeg": {
        "ffplay": "/default/ffplay",
        "ffprobe": "/default/ffprobe",
        "mplayer": "/default/mplayer"
    },
    "ws": "ws://default",
    "web": "http://default/song/get",
    "cache": "./default_cache"
}`

	tempFile := "test_env_config.json"
	err := os.WriteFile(tempFile, []byte(tempConfig), 0644)
	if err != nil {
		t.Fatal("Failed to create temp config file:", err)
	}
	defer os.Remove(tempFile) // clean up

	config, err := NewConfig(tempFile)
	if err != nil {
		t.Fatal("Failed to load config:", err)
	}

	// Environment variables should override file values
	if config.FFMpegConf.FFPlay != "/env/ffplay" {
		t.Errorf("Expected ffplay path to be overridden by env '/env/ffplay', got '%s'", config.FFMpegConf.FFPlay)
	}

	if config.WebSocketAPI != "ws://env-test" {
		t.Errorf("Expected WebSocket API to be overridden by env 'ws://env-test', got '%s'", config.WebSocketAPI)
	}
}

func TestConfigValidation(t *testing.T) {
	// Test with missing required fields
	tempConfig := `{
    "ffmpeg": {
        "ffplay": "",
        "ffprobe": ""
    },
    "ws": "",
    "web": "",
    "cache": ""
}`

	tempFile := "test_invalid_config.json"
	err := os.WriteFile(tempFile, []byte(tempConfig), 0644)
	if err != nil {
		t.Fatal("Failed to create temp config file:", err)
	}
	defer os.Remove(tempFile) // clean up

	_, err = NewConfig(tempFile)
	if err == nil {
		t.Error("Expected validation error for missing required fields, but got none")
	}
}
