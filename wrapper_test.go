package main

import (
	"runtime"
	"testing"
	"time"
)

func TestGetMediaInfo(t *testing.T) {
	handler := NewFFprobe(FFPROBE_BIN)
	//info, err := handler.GetMediaInfo(getLocalPath(MEDIA_PATH))
	info, err := handler.GetMediaInfo("http://www.google.com.hk")
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(info)

	duration, err := info.GetDuration()
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(duration)
}

func TestPlay(t *testing.T) {
	handler := NewFFplay(FFPLAY_BIN)
	handler.Play(getLocalPath(MEDIA_PATH), 10)

	runtime.GOMAXPROCS(runtime.NumCPU())

	done := make(chan bool, 0)

	time.AfterFunc(time.Second*3, func() {
		t.Log("time after 3s")
		handler.Stop()
	})

	time.AfterFunc(time.Second*5, func() {
		t.Log("time after 5s")
		handler.Play(getLocalPath(MEDIA_PATH), 20)
	})

	time.AfterFunc(time.Second*10, func() {
		t.Log("time after 10s")
		handler.Stop()

		done <- true
	})

	<-done
}
