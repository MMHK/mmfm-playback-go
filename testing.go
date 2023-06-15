package main

import (
	"path/filepath"
	"runtime"
)

const FFMPEG_BIN = "F:/green/ffmpeg/bin/ffmpeg.exe"
const FFPROBE_BIN = "F:/green/ffmpeg/bin/ffprobe.exe"
const FFPLAY_BIN = "F:/green/ffmpeg/bin/ffplay.exe"
const MPLAY_BIN = "F:/green/mplayer/mplayer.exe"
const MEDIA_PATH = "./temp/test.flac"
const WS_URL = "ws://localhost:8888/io/?EIO=3&transport=websocket"
const HTTP_API = "http://192.168.33.6:8888/song/get"

func getLocalPath(file string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), file)
}
