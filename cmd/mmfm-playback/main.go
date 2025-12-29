package main

import (
	"flag"
	"mmfm-playback-go/internal/config"
	"mmfm-playback-go/internal/logger"
	"mmfm-playback-go/internal/player"
	"runtime"
)

func main() {
	confPath := flag.String("c", "config.json", "config json file")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	conf, err := config.NewConfig(*confPath)
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	logger.Logger.Info("mmfm playback config: ", conf)

	mp := player.NewMusicPlayer(conf)
	logger.Logger.Info("mmfm playback start.")

	if err := mp.Start(); err != nil {
		logger.Logger.Error(err)
	}
}
