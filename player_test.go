package main

import "testing"

func TestNewMusicPlayer(t *testing.T) {
	err, conf := NewConfig(getLocalPath("./conf.json"))
	if err != nil {
		t.Error(err)
		return
	}
	NewMusicPlayer(conf)
}

func TestMusicPlayer_Start(t *testing.T) {
	err, conf := NewConfig(getLocalPath("./conf.json"))
	if err != nil {
		t.Error(err)
		return
	}
	player := NewMusicPlayer(conf)
	player.Start()
}
