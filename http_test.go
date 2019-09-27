package main

import "testing"

func TestLoadPlaylist(t *testing.T) {
	list, err := LoadPlaylist(HTTP_API)
	if err != nil {
		t.Error(err)
		return
	}

	for _, song := range list {
		t.Log(song)
	}
}
