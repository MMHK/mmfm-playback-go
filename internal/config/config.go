package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// FFmpegConfig holds FFmpeg related configuration
type FFmpegConfig struct {
	FFPlay  string `json:"ffplay"`
	FFProbe string `json:"ffprobe"`
	MPlayer string `json:"mplayer"`
}

// ScheduledAudio represents a scheduled audio playback configuration
type ScheduledAudio struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Schedule string `json:"schedule"` // Time in cron format or specific time format
}

// PlaybackConfig holds the main configuration for the playback service
type PlaybackConfig struct {
	FFMpegConf      *FFmpegConfig    `json:"ffmpeg"`
	WebSocketAPI    string           `json:"ws"`
	WebAPI          string           `json:"web"`
	CachePath       string           `json:"cache"`
	ScheduledAudios []ScheduledAudio `json:"scheduled_audios,omitempty"`
	configFile      string
}

// NewConfig creates a new configuration from file or environment variables
func NewConfig(filename string) (*PlaybackConfig, error) {
	c := &PlaybackConfig{
		FFMpegConf: &FFmpegConfig{},
		configFile: filename,
	}

	// Try to load from JSON file first
	if err := c.loadFromFile(filename); err != nil {
		fmt.Printf("Warning: Could not load config from file %s: %v\n", filename, err)
		fmt.Println("Falling back to environment variables...")
	}

	// Override with environment variables if present
	c.loadFromEnv()

	// Validate required fields
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return c, nil
}

// loadFromFile loads configuration from a JSON file
func (c *PlaybackConfig) loadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("could not open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(c); err != nil {
		return fmt.Errorf("could not decode config file: %w", err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func (c *PlaybackConfig) loadFromEnv() {
	// FFmpeg configuration
	if ffplay := os.Getenv("FFPLAY_PATH"); ffplay != "" {
		c.FFMpegConf.FFPlay = ffplay
	}
	if ffprobe := os.Getenv("FFPROBE_PATH"); ffprobe != "" {
		c.FFMpegConf.FFProbe = ffprobe
	}
	if mplayer := os.Getenv("MPLAYER_PATH"); mplayer != "" {
		c.FFMpegConf.MPlayer = mplayer
	}

	// API endpoints
	if wsAPI := os.Getenv("WEBSOCKET_API"); wsAPI != "" {
		c.WebSocketAPI = wsAPI
	}
	if webAPI := os.Getenv("WEB_API"); webAPI != "" {
		c.WebAPI = webAPI
	}

	// Cache path
	if cachePath := os.Getenv("CACHE_PATH"); cachePath != "" {
		c.CachePath = cachePath
	}

	// Support for legacy environment variable names as well
	if ws := os.Getenv("WS_API"); ws != "" {
		c.WebSocketAPI = ws
	}
	if web := os.Getenv("WEB_API_URL"); web != "" {
		c.WebAPI = web
	}
	if cache := os.Getenv("CACHE_DIR"); cache != "" {
		c.CachePath = cache
	}
}

// validate checks if required configuration fields are present
func (c *PlaybackConfig) validate() error {
	var missingFields []string

	if c.FFMpegConf.FFPlay == "" {
		missingFields = append(missingFields, "ffmpeg.ffplay")
	}
	if c.FFMpegConf.FFProbe == "" {
		missingFields = append(missingFields, "ffmpeg.ffprobe")
	}
	if c.WebSocketAPI == "" {
		missingFields = append(missingFields, "ws")
	}
	if c.WebAPI == "" {
		missingFields = append(missingFields, "web")
	}
	if c.CachePath == "" {
		missingFields = append(missingFields, "cache")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required configuration fields: %s", strings.Join(missingFields, ", "))
	}

	return nil
}

// Save saves the configuration to the JSON file
func (c *PlaybackConfig) Save() error {
	file, err := os.Create(c.configFile)
	if err != nil {
		return fmt.Errorf("could not create config file: %w", err)
	}
	defer file.Close()

	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("could not write config to file: %w", err)
	}

	return nil
}

// GetConfigPath returns the path of the configuration file
func (c *PlaybackConfig) GetConfigPath() string {
	return c.configFile
}
