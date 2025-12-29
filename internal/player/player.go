package player

import (
	"errors"
	"mmfm-playback-go/internal/cache"
	"mmfm-playback-go/internal/chat"
	"mmfm-playback-go/internal/config"
	"mmfm-playback-go/internal/logger"
	"mmfm-playback-go/pkg/types"
	"time"
)

var Logger = logger.Logger

// Player interface defines the music player functionality
type Player interface {
	Start() error
	Play(song *types.Song, second int) error
	Next()
	GetSongInPlayList(index int) (*types.Song, error)
}

// MusicPlayer is the main music player implementation
type MusicPlayer struct {
	Conf         *config.PlaybackConfig
	player       *Mplayer
	probe        *FFprobe
	playlist     []*types.Song
	currentIndex float64
	chat         *chat.ChatClient
	currentSong  *types.Song
	cache        *cache.FileCache
	pauseFlag    bool
}

// NewMusicPlayer creates a new music player instance
func NewMusicPlayer(conf *config.PlaybackConfig) *MusicPlayer {
	return &MusicPlayer{
		Conf:         conf,
		player:       NewMplayer(conf.FFMpegConf.MPlayer),
		probe:        NewFFprobe(conf.FFMpegConf.FFProbe),
		playlist:     make([]*types.Song, 0),
		currentIndex: 0,
		pauseFlag:    true,
		cache:        cache.NewFileCache(conf.CachePath),
		chat:         chat.NewChatClient(conf.WebSocketAPI),
	}
}

// Start initializes and starts the music player
func (mp *MusicPlayer) Start() error {
	retryCounter := 0
start:
	list, err := LoadPlaylist(mp.Conf.WebAPI)
	retryCounter++
	if err != nil {
		Logger.Error(err)
		if retryCounter < 10 {
			Logger.Error("retry connect!")
			time.Sleep(time.Second * 2)
			goto start
		}
		return err
	}
	mp.playlist = list

	go mp.cache.Clean(mp.playlist)

	if len(mp.playlist) > 0 {
		song, err := mp.GetSongInPlayList(int(mp.currentIndex))
		if err != nil {
			Logger.Error(err)
			return err
		}
		go func() {
			err := mp.Play(song, 0)
			if err != nil {
				Logger.Error(err)
				mp.Next()
			}
		}()
	}

	go mp.TrackPlaying()
	mp.Listen()

	return nil
}

// Listen handles incoming chat messages
func (mp *MusicPlayer) Listen() error {
	listener, err := mp.chat.Listen()
	if err != nil {
		Logger.Error(err)
		return err
	}

	for {
		msg := <-listener
		switch msg.Command {
		case "player.play":
			if len(msg.Params) > 1 {
				mp.pauseFlag = true

				index, ok := msg.Params[1].(float64)
				if ok && index != mp.currentIndex {
					mp.currentIndex = index
				} else {
					mp.currentIndex = 0
				}

				song, err := mp.GetSongInPlayList(int(mp.currentIndex))
				if err != nil {
					Logger.Error(err)
					mp.Next()
					break
				}
				mp.player.Stop()
				err = mp.Play(song, 0)
				if err != nil {
					Logger.Error(err)
					mp.Next()
					break
				}
			}
			break

		case "player.continue":
			mp.pauseFlag = false
			go mp.Play(mp.currentSong, int(mp.currentSong.Index))
			break

		case "player.pause":
			Logger.Debug("pause song", mp.currentSong.Name)
			mp.pauseFlag = true
			mp.player.Stop()
			mp.FirePause()
			break

		case "player.current":
			if mp.chat != nil {
				if mp.pauseFlag {
					mp.FirePause()
				} else {
					mp.FirePlaying()
				}
			}
			break

		case "update":
			Logger.Debug("update playlist")
			list, err := LoadPlaylist(mp.Conf.WebAPI)
			if err != nil {
				Logger.Error(err)
				break
			}
			mp.playlist = list
			break
		}
	}

	return nil
}

// FirePause sends a pause event
func (mp *MusicPlayer) FirePause() {
	if mp.chat != nil && mp.currentSong != nil {
		mp.currentSong.URL = mp.currentSong.GetURL()
		mp.chat.SendEvent("msg", &chat.MessageArgs{
			Command: "player.pause",
			Params: []interface{}{
				mp.currentSong,
				mp.currentIndex,
				mp.currentSong.Index,
				mp.currentSong.Duration,
			},
		})
	}
}

// FirePlaying sends a playing event
func (mp *MusicPlayer) FirePlaying() {
	if mp.chat != nil && mp.currentSong != nil {
		mp.currentSong.URL = mp.currentSong.GetURL()
		if !mp.pauseFlag {
			mp.currentSong.Index = mp.currentSong.Index + 1
		}
		mp.chat.SendEvent("msg", &chat.MessageArgs{
			Command: "player.playing",
			Params: []interface{}{
				mp.currentSong,
				mp.currentIndex,
				mp.currentSong.Index,
				mp.currentSong.Duration,
			},
		})
	}
}

// TrackPlaying continuously sends playing events
func (mp *MusicPlayer) TrackPlaying() {
	for {
		if !mp.pauseFlag {
			mp.FirePlaying()
		}
		time.Sleep(time.Second * 1)
	}
}

// LoadPlaylist loads a playlist from a web API
func LoadPlaylist(apiURL string) ([]*types.Song, error) {
	// This is a placeholder implementation
	// In a real implementation, you would make an HTTP request to apiURL
	// and parse the response to get a list of songs
	Logger.Info("Loading playlist from", apiURL)
	return []*types.Song{}, nil
}

// Play plays a song from a specific time
func (mp *MusicPlayer) Play(song *types.Song, second int) error {
	Logger.Debug("play song", song.Name)
	url := mp.cache.Cache(song.GetURL())

	info, err := mp.probe.GetMediaInfo(url)
	if err != nil {
		Logger.Error(err)
		mp.pauseFlag = true
		return err
	}
	song.Index = float64(second)
	duration, err := info.GetDuration()
	if err != nil {
		Logger.Error(err)
		mp.pauseFlag = true
		return err
	}
	Logger.Debug(duration)

	song.Duration = duration
	finish, err := mp.player.Play(url, second)
	mp.pauseFlag = false
	mp.currentSong = song

	go func() {
		<-finish
		if !mp.pauseFlag {
			mp.Next()
		}
	}()

	return nil
}

// GetSongInPlayList retrieves a song from the playlist by index
func (mp *MusicPlayer) GetSongInPlayList(index int) (*types.Song, error) {
	if index >= 0 && index < len(mp.playlist) {
		return mp.playlist[index], nil
	}

	if len(mp.playlist) > 0 {
		mp.currentIndex = 0
		return mp.playlist[0], nil
	}

	return nil, errors.New("song not found")
}

// Next plays the next song in the playlist
func (mp *MusicPlayer) Next() {
	index := mp.currentIndex + 1
	if index > float64(len(mp.playlist)-1) {
		index = 0
	}
	mp.Play(mp.playlist[int(index)], 0)
	mp.currentIndex = index
}
