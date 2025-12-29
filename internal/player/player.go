package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"mmfm-playback-go/internal/cache"
	"mmfm-playback-go/internal/chat"
	"mmfm-playback-go/internal/config"
	"mmfm-playback-go/internal/logger"
	"mmfm-playback-go/pkg/types"
	"net/http"
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
	// Add fields for scheduled audio playback
	scheduledAudioPlaying bool
	originalPaused        bool
}

// NewMusicPlayer creates a new music player instance
func NewMusicPlayer(conf *config.PlaybackConfig) *MusicPlayer {
	player := &MusicPlayer{
		Conf:         conf,
		player:       NewMplayer(conf.FFMpegConf.MPlayer),
		probe:        NewFFprobe(conf.FFMpegConf.FFProbe),
		playlist:     make([]*types.Song, 0),
		currentIndex: 0,
		pauseFlag:    true,
		cache:        cache.NewFileCache(conf.CachePath),
		chat:         chat.NewChatClient(conf.WebSocketAPI),
	}

	// Initialize scheduled audio handling if scheduled audios are configured
	if len(conf.ScheduledAudios) > 0 {
		go player.handleScheduledAudios()
	}

	return player
}

// handleScheduledAudios manages scheduled audio playback
func (mp *MusicPlayer) handleScheduledAudios() {
	for {
		// Check for scheduled audios that should play now
		for _, scheduledAudio := range mp.Conf.ScheduledAudios {
			if mp.isTimeToPlay(scheduledAudio.Schedule) {
				// Play the scheduled audio
				mp.playScheduledAudio(scheduledAudio)
			}
		}
		// Check every 30 seconds for scheduled audios
		time.Sleep(30 * time.Second)
	}
}

// isTimeToPlay checks if the current time matches the schedule
func (mp *MusicPlayer) isTimeToPlay(schedule string) bool {
	if mp.scheduledAudioPlaying || mp.currentSong == nil {
		return false
	}
	// For now, we'll implement a simple time format check
	// Expected format: "HH:MM" for daily scheduling
	if len(schedule) == 5 && string(schedule[2]) == ":" {
		currentTime := time.Now()
		currentHour := currentTime.Hour()
		currentMinute := currentTime.Minute()

		var hour, minute int
		if _, err := fmt.Sscanf(schedule, "%d:%d", &hour, &minute); err != nil {
			return false
		}

		return currentHour == hour && currentMinute == minute
	}

	// TODO: Implement cron-like scheduling support for more complex schedules
	Logger.Debug("Schedule format not yet supported:", schedule)
	return false
}

// playScheduledAudio handles playing a scheduled audio, pausing current playback
func (mp *MusicPlayer) playScheduledAudio(scheduledAudio config.ScheduledAudio) {
	Logger.Infof("Playing scheduled audio: %s at %s", scheduledAudio.Name, scheduledAudio.URL)

	// Check if we're already playing a scheduled audio
	if mp.scheduledAudioPlaying {
		Logger.Debug("Already playing a scheduled audio, skipping:", scheduledAudio.Name)
		return
	}

	// Pause current playback and save state
	mp.scheduledAudioPlaying = true
	mp.originalPaused = mp.pauseFlag

	if mp.currentSong != nil {
		Logger.Debug("Pausing current song:", mp.currentSong.Name)
		mp.Pause()
	}

	// Create a temporary song object for the scheduled audio
	tempSong := &types.Song{
		Name:     scheduledAudio.Name,
		URL:      scheduledAudio.URL,
		Duration: 0, // Will be updated after probing
		Index:    0,
	}

	// Play the scheduled audio
	go func() {
		err := mp.playWithoutInterrupt(tempSong, 0)
		if err != nil {
			Logger.Error("Error playing scheduled audio:", err)
		}
		// After scheduled audio finishes, resume original playback
		mp.resumeOriginalPlayback()
	}()
}

// playWithoutInterrupt plays an audio without triggering normal playback events
func (mp *MusicPlayer) playWithoutInterrupt(song *types.Song, second int) error {
	Logger.Debug("Playing scheduled audio without interrupting normal flow", song.Name)
	url := mp.cache.Cache(song.GetURL())

	info, err := mp.probe.GetMediaInfo(url)
	if err != nil {
		Logger.Error(err)
		return err
	}
	song.Index = float64(second)
	duration, err := info.GetDuration()
	if err != nil {
		Logger.Error(err)
		return err
	}
	Logger.Debug("Scheduled audio duration:", duration)

	song.Duration = duration
	finish, err := mp.player.Play(url, second)
	if err != nil {
		return err
	}

	// Wait for the audio to finish
	<-finish
	return nil
}

// resumeOriginalPlayback restores the original playback after scheduled audio
func (mp *MusicPlayer) resumeOriginalPlayback() {
	Logger.Info("Resuming original playback after scheduled audio")

	// Reset scheduled audio flag
	mp.scheduledAudioPlaying = false

	// If original playback was not paused, resume it
	if !mp.originalPaused {
		if mp.currentSong != nil {
			// Resume from the saved position
			go func() {
				err := mp.Play(mp.currentSong, int(mp.currentSong.Index))
				if err != nil {
					Logger.Error("Error resuming original playback:", err)
					// If resume fails, continue with normal playback
					mp.Next()
				}
			}()
		} else {
			// If no original song, just continue with next in playlist
			mp.Next()
		}
	} else {
		// Original was paused, so keep it paused
		mp.pauseFlag = true
		mp.FirePause()
	}
}

// Pause pauses the current playback
func (mp *MusicPlayer) Pause() {
	Logger.Debug("Pausing song", mp.currentSong.Name)
	mp.pauseFlag = true
	mp.player.Stop()
	mp.FirePause()
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
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var playlist []*types.Song
	err = json.NewDecoder(resp.Body).Decode(&playlist)
	if err != nil {
		return nil, err
	}
	Logger.Debug("Loaded playlist:", playlist)
	return playlist, nil
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
	mp.currentSong.Duration = duration
	logger.Logger.Infof("playing song %s, duration %f, start %d", song.Name, duration, second)
	mp.FirePlaying()

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
