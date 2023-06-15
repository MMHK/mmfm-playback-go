package main

import (
	"errors"
	"time"
)

type MusicPlayer struct {
	Conf         *PlaybackConfig
	player       *Mplayer
	probe        *FFprobe
	playlist     []*Song
	currentIndex float64
	chat         *ChatClient
	currentSong  *Song
	cache        *FileCache
	pauseFlag    bool
}

func NewMusicPlayer(conf *PlaybackConfig) *MusicPlayer {

	return &MusicPlayer{
		Conf:         conf,
		player:       NewMplayer(conf.FFMpegConf.MPlayer),
		probe:        NewFFprobe(conf.FFMpegConf.FFProbe),
		playlist:     make([]*Song, 0),
		currentIndex: 0,
		pauseFlag:    true,
		cache:        NewFileCache(conf.CachePath),
		chat:         NewChatClient(conf.WebSocketAPI),
	}
}

func (this *MusicPlayer) Start() error {

	retryCounter := 0
start:
	list, err := LoadPlaylist(this.Conf.WebAPI)
	retryCounter++
	if err != nil {
		log.Error(err)
		if retryCounter < 10 {
			log.Error("retry connect!")
			time.Sleep(time.Second * 2)
			goto start
		}
		return err
	}
	this.playlist = list

	go this.cache.Clean(this.playlist)

	if len(this.playlist) > 0 {
		song, err := this.GetSongInPlayList(int(this.currentIndex))
		if err != nil {
			log.Error(err)
			return err
		}
		go func() {
			err := this.Play(song, 0)
			if err != nil {
				log.Error(err)
				this.Next()
			}
		}()
	}

	go this.TrackPlaying()
	this.Listen()

	return nil
}

func (this *MusicPlayer) Listen() error {
	listener, err := this.chat.Listen()
	if err != nil {
		log.Error(err)
		return err
	}

	for {
		msg := <-listener
		switch msg.Command {
		case EVENT_PLAY:
			if len(msg.Params) > 1 {
				this.pauseFlag = true

				index, ok := msg.Params[1].(float64)
				if ok && index != this.currentIndex {
					this.currentIndex = index
				} else {
					this.currentIndex = 0
				}

				song, err := this.GetSongInPlayList(int(this.currentIndex))
				if err != nil {
					log.Error(err)
					this.Next()
					break
				}
				this.player.Stop()
				err = this.Play(song, 0)
				if err != nil {
					log.Error(err)
					this.Next()
					break
				}
			}
			break

		case EVENT_CONTINUE:
			this.pauseFlag = false
			go this.Play(this.currentSong, int(this.currentSong.Index))
			break

		case EVENT_PAUSE:
			log.Debug("pause song", this.currentSong.Name)
			this.pauseFlag = true
			this.player.Stop()
			this.FirePause()
			break

		case EVENT_CURRENT:
			if this.chat != nil {
				if this.pauseFlag {
					this.FirePause()
				} else {
					this.FirePlaying()
				}
			}
			break

		case EVENT_UPDATE:
			log.Debug("update playlist")
			list, err := LoadPlaylist(this.Conf.WebAPI)
			if err != nil {
				log.Error(err)
				break
			}
			this.playlist = list
			break
		}
	}

	return nil
}

func (this *MusicPlayer) GetSongInPlayList(index int) (*Song, error) {
	if index >= 0 && index < len(this.playlist) {
		return this.playlist[index], nil
	}

	if len(this.playlist) > 0 {
		this.currentIndex = 0
		return this.playlist[0], nil
	}

	return nil, errors.New("song not found")
}

func (this *MusicPlayer) Play(song *Song, second int) error {
	log.Debug("play song", song.Name)
	url := this.cache.Cache(song.GetURL())

	info, err := this.probe.GetMediaInfo(url)
	if err != nil {
		log.Error(err)
		this.pauseFlag = true
		return err
	}
	song.Index = float64(second)
	duration, err := info.GetDuration()
	if err != nil {
		log.Error(err)
		this.pauseFlag = true
		return err
	}
	log.Debug(duration)

	song.Duration = duration
	finish, err := this.player.Play(url, second)
	this.pauseFlag = false
	this.currentSong = song

	go func() {
		<-finish
		if !this.pauseFlag {
			this.Next()
		}
	}()

	return nil
}

func (this *MusicPlayer) FirePause() {
	if this.chat != nil && this.currentSong != nil {
		this.currentSong.URL = this.currentSong.GetURL()
		this.chat.SendEvent(CHAT_EVENT_MESSAGE, &MessageArgs{
			Command: EVENT_PAUSE,
			Params: []interface{}{
				this.currentSong,
				this.currentIndex,
				this.currentSong.Index,
				this.currentSong.Duration,
			},
		})
	}
}

func (this *MusicPlayer) FirePlaying() {
	if this.chat != nil && this.currentSong != nil {
		this.currentSong.URL = this.currentSong.GetURL()
		if !this.pauseFlag {
			this.currentSong.Index = this.currentSong.Index + 1
		}
		this.chat.SendEvent(CHAT_EVENT_MESSAGE, &MessageArgs{
			Command: EVENT_PLAYING,
			Params: []interface{}{
				this.currentSong,
				this.currentIndex,
				this.currentSong.Index,
				this.currentSong.Duration,
			},
		})
	}
}

func (this *MusicPlayer) TrackPlaying() {
	for {
		if !this.pauseFlag {
			this.FirePlaying()
		}
		time.Sleep(time.Second * 1)
	}
}

func (this *MusicPlayer) Next() {
	index := this.currentIndex + 1
	if index > float64(len(this.playlist)-1) {
		index = 0
	}
	this.Play(this.playlist[int(index)], 0)
	this.currentIndex = index
}
