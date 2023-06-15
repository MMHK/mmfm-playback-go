// config
package main

import (
	"encoding/json"
	"os"
)

type FFmpegConfig struct {
	FFPlay  string `json:"ffplay"`
	FFProbe string `json:"ffprobe"`
	MPlayer string `json:"mplayer"`
}

type PlaybackConfig struct {
	FFMpegConf   *FFmpegConfig `json:"ffmpeg"`
	WebSocketAPI string        `json:"ws"`
	WebAPI       string        `json:"web"`
	CachePath    string        `json:"cache"`
	sava_file    string
}

func NewConfig(filename string) (err error, c *PlaybackConfig) {
	c = &PlaybackConfig{}
	c.sava_file = filename
	err = c.load(filename)
	if err != nil {
		return err, nil
	}
	return nil, c
}

func (c *PlaybackConfig) load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Error(err)
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		log.Error(err)
	}
	return err
}

func (c *PlaybackConfig) Save() error {
	file, err := os.Create(c.sava_file)
	if err != nil {
		log.Error(err)
		return err
	}
	defer file.Close()
	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		log.Error(err)
	}
	return err
}
