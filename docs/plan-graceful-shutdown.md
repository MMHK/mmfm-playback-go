# 規劃：平滑退出（信號處理）+ 服務運行信息輸出

## 目標

1. 在 `cmd/mmfm-playback/main.go` 加入信號監聽（SIGINT/SIGTERM），實現平滑退出
2. 服務啟動時輸出運行信息（PID、Go 版本、OS/Arch、配置路徑等）
3. 為 `MusicPlayer` 新增 `Stop()` 方法，支援乾淨關閉所有 goroutine 與外部連線
4. 在 `internal/chat/chat.go` 加入連線重試機制，防止 MMFM 服務尚未啟動時連線失敗

## 任務

- [x] 1. 在 `internal/player/player.go` 為 `MusicPlayer` 新增 `stopCh chan struct{}` 欄位，用於通知 goroutine 退出
- [x] 2. 在 `NewMusicPlayer()` 初始化 `stopCh`
- [x] 3. 新增 `MusicPlayer.Stop()` 方法：關閉 `stopCh`、停止 mplayer、關閉 chat client
- [x] 4. 修改 `handleScheduledAudios()`：使用 `select` + `stopCh` 支援退出
- [x] 5. 修改 `TrackPlaying()`：使用 `select` + `stopCh` 支援退出
- [x] 6. 修改 `Listen()`：在 `for` 迴圈中使用 `select` + `stopCh` 支援退出
- [x] 7. 修改 `internal/chat/chat.go` `Listen()`：加入連線重試機制，失敗時重試直到成功或收到停止信號
- [x] 8. 修改 `cmd/mmfm-playback/main.go`：加入 `os/signal` 監聽 SIGINT/SIGTERM，收到信號時呼叫 `mp.Stop()`
- [x] 9. 修改 `cmd/mmfm-playback/main.go`：啟動時輸出服務運行信息（PID、Go 版本、OS/Arch、配置路徑、啟動時間）
- [x] 10. 執行 `go vet ./...` 確認無新警告（`player.go:352` unreachable code 為已知問題）
- [x] 11. 執行 `go test -v ./...` 確認測試通過（全部 PASS）

## 詳細設計

### 1. MusicPlayer 新增 stopCh 欄位

```go
type MusicPlayer struct {
    // ... 現有欄位
    stopCh chan struct{}  // 新增：用於通知 goroutine 退出
}
```

### 2. NewMusicPlayer 初始化 stopCh

```go
func NewMusicPlayer(conf *config.PlaybackConfig) *MusicPlayer {
    player := &MusicPlayer{
        // ... 現有初始化
        stopCh: make(chan struct{}),  // 新增
    }
    // ...
}
```

### 3. MusicPlayer.Stop() 方法

```go
func (mp *MusicPlayer) Stop() {
    slog.Info("stopping music player")
    close(mp.stopCh)
    mp.player.Stop()
    if mp.chat != nil {
        mp.chat.Close()
    }
    slog.Info("music player stopped")
}
```

### 4. handleScheduledAudios 支援退出

```go
func (mp *MusicPlayer) handleScheduledAudios() {
    for {
        select {
        case <-mp.stopCh:
            slog.Debug("scheduled audio handler stopped")
            return
        case <-time.After(30 * time.Second):
            for _, scheduledAudio := range mp.Conf.ScheduledAudios {
                if mp.isTimeToPlay(scheduledAudio.Schedule) {
                    mp.playScheduledAudio(scheduledAudio)
                }
            }
        }
    }
}
```

### 5. TrackPlaying 支援退出

```go
func (mp *MusicPlayer) TrackPlaying() {
    for {
        select {
        case <-mp.stopCh:
            slog.Debug("track player stopped")
            return
        case <-time.After(time.Second):
            if !mp.pauseFlag {
                mp.FirePlaying()
            }
        }
    }
}
```

### 6. Listen 支援退出

將 `msg := <-listener` 改為 `select`，並傳入 `stopCh` 給 `chat.Listen()`：

```go
func (mp *MusicPlayer) Listen() error {
    listener, err := mp.chat.Listen(mp.stopCh)
    if err != nil {
        slog.Error("chat listen failed", "error", err)
        return err
    }

    for {
        select {
        case <-mp.stopCh:
            slog.Info("listener stopped")
            return nil
        case msg := <-listener:
            // ... 現有 switch 邏輯
        }
    }
}
```

### 7. ChatClient 連線重試機制

修改 `ChatClient.Listen()` 方法，加入連線重試邏輯。需接受 `stopCh` 參數以便在收到停止信號時退出重試迴圈：

```go
func (cc *ChatClient) Listen(stopCh <-chan struct{}) (chan *MessageArgs, error) {
    retryCounter := 0
    const maxRetries = 30
    const retryInterval = 2 * time.Second

connect:
    for {
        select {
        case <-stopCh:
            return nil, errors.New("connection cancelled")
        default:
        }

        err := cc.Connect()
        if err == nil {
            break
        }

        retryCounter++
        if retryCounter >= maxRetries {
            return nil, fmt.Errorf("connect failed after %d retries: %w", maxRetries, err)
        }

        slog.Error("connect failed, retrying",
            "attempt", retryCounter,
            "max_retries", maxRetries,
            "error", err,
        )

        select {
        case <-stopCh:
            return nil, errors.New("connection cancelled during retry")
        case <-time.After(retryInterval):
            goto connect
        }
    }

    // ... 現有事件註冊邏輯
    return cc.listener, nil
}
```

**注意事項：**
- 重試間隔固定 2 秒，最多重試 30 次（約 60 秒）
- 每次重試前檢查 `stopCh`，確保可被中斷
- 需同步修改 `player.go` 的 `Listen()` 呼叫，傳入 `mp.stopCh`

### 8. main.go 信號處理

```go
func main() {
    confPath := flag.String("c", "config.json", "config json file")
    flag.Parse()

    runtime.GOMAXPROCS(runtime.NumCPU())
    _ = godotenv.Overload()
    logger.Init()

    // 輸出服務運行信息
    printServiceInfo(*confPath)

    conf, err := config.NewConfig(*confPath)
    if err != nil {
        slog.Error("error", "error", err)
        return
    }
    slog.Info("mmfm playback config", "config", conf)

    mp := player.NewMusicPlayer(conf)

    // 設置信號處理
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
```

### 9. 服務運行信息輸出

```go
func printServiceInfo(configPath string) {
    slog.Info("mmfm-playback-go starting",
        "pid", os.Getpid(),
        "go_version", runtime.Version(),
        "os_arch", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
        "num_cpu", runtime.NumCPU(),
        "config_path", configPath,
        "start_time", time.Now().Format(time.RFC3339),
    )
}
```

## 受影響的檔案

| 檔案 | 改動說明 |
|------|---------|
| `internal/player/player.go` | 新增 `stopCh` 欄位、`Stop()` 方法；修改 `handleScheduledAudios()`、`TrackPlaying()`、`Listen()` 支援退出 |
| `internal/chat/chat.go` | `Listen()` 加入重試機制 + 接受 `stopCh` 參數 |
| `cmd/mmfm-playback/main.go` | 新增信號處理、服務信息輸出函式 |

## 風險

1. **`close(mp.stopCh)` 不可重複呼叫**：若 `Stop()` 被呼叫兩次會 panic。可考慮使用 `sync.Once` 保護，或檢查 channel 是否已關閉。
2. **`Listen()` 返回後 `Start()` 也會返回**：目前 `Start()` 最後一行是 `mp.Listen()`，信號處理後 `Listen()` 返回 nil，`Start()` 也會正常結束。
3. **goroutine 退出時間**：`time.Sleep(30 * time.Second)` 改為 `time.After` 後，最長需等待 30 秒才能退出。可接受，因為信號處理是非阻塞的。
4. **chat.Close() 可能阻塞**：需確認 `socketio.Socket.Disconnect()` 是否會阻塞。若會阻塞，需設置超時。

## 驗證步驟

1. 執行 `make build` 確認編譯通過
2. 執行 `go vet ./...` 確認無警告
3. 執行 `make test` 確認測試通過
4. 手動測試：啟動服務後按 Ctrl+C，觀察是否有平滑退出日誌
5. 確認啟動時有輸出 PID、Go 版本、OS/Arch 等信息
