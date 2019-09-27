package main

import (
	"path/filepath"
	"runtime"
)

const FFMPEG_BIN = "F:/grean/ffmpeg/bin/ffmpeg.exe"
const FFPROBE_BIN = "F:/grean/ffmpeg/bin/ffprobe.exe"
const FFPLAY_BIN = "F:/grean/ffmpeg/bin/ffplay.exe"
const MEDIA_PATH = "./testdata/2355975676.mp3"
const WS_URL = "ws://localhost:8888/io/?EIO=3&transport=websocket"
const HTTP_API = "http://192.168.33.6:8888/song/get"

func getLocalPath(file string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), file)
}
