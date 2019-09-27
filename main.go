package main

import (
	"flag"
	"runtime"
)

func main() {
	conf_path := flag.String("c", "conf.json", "config json file")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())


	err, conf := NewConfig(*conf_path)
	if err != nil {
		log.Error(err)
		return
	}
	player := NewMusicPlayer(conf)
	log.Info("mmfm playback start.")
	player.Start()
}
