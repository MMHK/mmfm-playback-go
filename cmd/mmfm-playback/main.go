package main

import (
	"flag"
	"fmt"
	"log/slog"
	"mmfm-playback-go/internal/config"
	"mmfm-playback-go/internal/logger"
	"mmfm-playback-go/internal/player"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func printServiceInfo(configPath string) {
	slog.Info("mmfm-playback-go starting",
		"version", version,
		"commit", commit,
		"date", date,
		"pid", os.Getpid(),
		"go_version", runtime.Version(),
		"os_arch", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		"num_cpu", runtime.NumCPU(),
		"config_path", configPath,
		"start_time", time.Now().Format(time.RFC3339),
	)
}

func main() {
	confPath := flag.String("c", "config.json", "config json file")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	_ = godotenv.Overload()

	logger.Init()

	printServiceInfo(*confPath)

	conf, err := config.NewConfig(*confPath)
	if err != nil {
		slog.Error("error", "error", err)
		return
	}
	slog.Info("mmfm playback config", "config", conf)

	mp := player.NewMusicPlayer(conf)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		slog.Info("received signal, shutting down", "signal", sig)
		mp.Stop()
	}()

	slog.Info("mmfm playback start")

	if err := mp.Start(); err != nil {
		slog.Error("error", "error", err)
	}
}
