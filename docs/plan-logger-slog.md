# 規劃：Logger 更換為 log/slog + 完善 Debug Log

## 目標

將現有 `github.com/op/go-logging` 替換為 Go 標準庫 `log/slog`，全面改用結構化日誌（JSON 格式），並透過 `LOG_LEVEL` 環境變數控制日誌等級，使 Debug 日誌可按需啟用。

## 任務

- [x] 1. 重寫 `internal/logger/logger.go`：使用 `log/slog` 的 `JSONHandler`，讀取 `LOG_LEVEL` 環境變數設定等級（預設 `INFO`）
- [x] 2. 更新 `internal/logger/logger_test.go`：驗證 slog logger 初始化及等級設定
- [x] 3. 遷移 `cmd/mmfm-playback/main.go`：改用 `slog` 結構化呼叫
- [x] 4. 遷移 `internal/cache/cache.go`：改用 `slog` 結構化呼叫（11 處）
- [x] 5. 遷移 `internal/chat/chat.go`：改用 `slog` 結構化呼叫，並完善通訊 debug log（詳見下方 6a）
- [x] 5a. 完善 `internal/chat/chat.go` 通訊 debug log：新增連線、斷線、重連、送事件、收訊息解析等除錯資訊
- [x] 6. 遷移 `internal/player/player.go`：改用 `slog` 結構化呼叫（29 處），移除 `var Logger = logger.Logger` 別名
- [x] 7. 遷移 `internal/probe/probe.go`：改用 `slog` 結構化呼叫（1 處）
- [x] 8. 執行 `go mod tidy` + `go mod vendor` 移除 `github.com/op/go-logging` 依賴
- [x] 9. 執行 `go vet ./...` 和 `go test -mod=readonly -v ./...` 驗證正確性
- [x] 10. 更新 `AGENTS.md`：移除舊 logger 相關說明，記錄新的 `LOG_LEVEL` 環境變數

## 詳細設計

### 1. `internal/logger/logger.go` 新實作

```go
package logger

import (
    "log/slog"
    "os"
    "strings"
)

func Init() {
    level := slog.LevelInfo
    if env := os.Getenv("LOG_LEVEL"); env != "" {
        switch strings.ToUpper(env) {
        case "DEBUG":
            level = slog.LevelDebug
        case "INFO":
            level = slog.LevelInfo
        case "WARN":
            level = slog.LevelWarn
        case "ERROR":
            level = slog.LevelError
        }
    }

    handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
        Level: level,
    })
    slog.SetDefault(slog.New(handler))
}
```

- 使用 `slog.NewJSONHandler` 輸出 JSON 格式至 stderr
- 透過 `LOG_LEVEL` 環境變數控制等級，支援 `DEBUG`/`INFO`/`WARN`/`ERROR`
- 預設 `INFO`
- 提供 `Init()` 函式供 main.go 在啟動時呼叫

### 2. 結構化日誌遷移對照

| 現有寫法 | 新寫法 |
|---------|--------|
| `logger.Logger.Error(err)` | `slog.Error("error", "error", err)` |
| `logger.Logger.Info("cache hint ", path)` | `slog.Info("cache hit", "path", path)` |
| `logger.Logger.Debug("begin cache music file")` | `slog.Debug("begin cache music file")` |
| `logger.Logger.Error("url return:", resp.StatusCode)` | `slog.Error("url return", "status_code", resp.StatusCode)` |
| `Logger.Infof("Playing scheduled audio: %s at %s", name, url)` | `slog.Info("playing scheduled audio", "name", name, "url", url)` |
| `Logger.Debug(duration)` | `slog.Debug("duration", "duration", duration)` |

### 3. `cmd/mmfm-playback/main.go` 初始化流程

```go
func main() {
    confPath := flag.String("c", "config.json", "config json file")
    flag.Parse()

    runtime.GOMAXPROCS(runtime.NumCPU())
    _ = godotenv.Overload()

    logger.Init()   // 初始化 slog（在 .env 載入之後）

    // ... 後續使用 slog.Info/Error 等
}
```

### 4. player.go 特殊處理

- 移除 `var Logger = logger.Logger` 別名（第 16 行）
- 移除 `"mmfm-playback-go/internal/logger"` import（若不再需要）
- 但 main.go 仍需 import logger 來呼叫 `logger.Init()`
- 其他檔案若只用 slog 則可移除 logger import

### 5. LOG_LEVEL 環境變數

| 值 | 效果 |
|----|------|
| `DEBUG` | 輸出所有等級（Debug/Info/Warn/Error） |
| `INFO` | 預設值，輸出 Info/Warn/Error |
| `WARN` | 輸出 Warn/Error |
| `ERROR` | 僅輸出 Error |

### 5a. `internal/chat/chat.go` 通訊 Debug Log 完善

#### 現有问题

| 問題 | 說明 |
|------|------|
| `Connect()` 無連線目標 log | 無法確認連線的 URL 和 path |
| 缺少 Socket.IO 生命周期事件 | `connect_error`、`reconnect`、`reconnect_attempt`、`reconnect_failed` 皆未監聽 |
| `SendEvent()` 無成功 log | 只有錯誤時 log，無法追蹤正常通訊（需避免過於頻繁，`TrackPlaying` 每秒送一次） |
| listener channel 可能靜默阻塞 | `cc.listener <-` 在 buffer 滿時會阻塞，無保護也無 log |
| 收到訊息只 log raw string | 沒有 log 解析後的 `command` 類型 |
| `Close()` 無 log | 斷線時無跡可尋 |

#### 新增的 Debug Log

| 位置 | 事件 | Log 內容 |
|------|------|---------|
| `Connect()` | 連線嘗試 | `slog.Debug("connecting", "url", cc.url, "path", "/io")` |
| `Connect()` | 連線成功 | `slog.Info("connected", "url", cc.url)` （取代現有 "connected"） |
| `Listen()` | 註冊 connect_error handler | 監聽 `connect_error` 事件 |
| `connect_error` | 連線錯誤 | `slog.Error("connect error", "error", err)` |
| `Listen()` | 註冊 reconnect handler | 監聽 `reconnect` 事件 |
| `reconnect` | 重連成功 | `slog.Info("reconnected", "attempt", attemptNumber)` |
| `Listen()` | 註冊 reconnect_attempt handler | 監聽 `reconnect_attempt` 事件 |
| `reconnect_attempt` | 重連嘗試 | `slog.Debug("reconnect attempt", "attempt", attemptNumber)` |
| `Listen()` | 註冊 reconnect_failed handler | 監聽 `reconnect_failed` 事件 |
| `reconnect_failed` | 重連失敗 | `slog.Error("reconnect failed")` |
| `disconnect` | 斷線原因 | `slog.Info("disconnected", "reason", reason)` （帶 reason 參數） |
| `msg` handler | 收到訊息 | `slog.Debug("received message", "raw", sourceParams)` |
| `msg` handler | 解析結果 | `slog.Debug("parsed message", "command", params.Command, "param_count", len(params.Params))` |
| `msg` handler | channel 滿時 | 改用 `select { case cc.listener <- msg: default: slog.Warn("listener channel full, message dropped") }` |
| `SendEvent()` | 送出事件 | `slog.Debug("sending event", "event", eventName, "command", params.Command)` |
| `SendEvent()` | 送出成功 | `slog.Debug("event sent", "event", eventName)` |
| `Close()` | 關閉連線 | `slog.Info("closing chat client")` |
| `ParseMessageArgs()` | JSON 解析失敗 | `slog.Error("parse message failed", "error", err, "raw", source)` （需將 source 傳入） |
| `GetPlayingEvent()` | recover panic | `slog.Error("panic in GetPlayingEvent", "error", fmt.Sprintf("%v", r))` |

## 受影響的檔案

| 檔案 | 改動說明 |
|------|---------|
| `internal/logger/logger.go` | 完全重寫，改用 `log/slog` |
| `internal/logger/logger_test.go` | 更新測試 |
| `cmd/mmfm-playback/main.go` | 改用 `slog` + 呼叫 `logger.Init()` |
| `internal/cache/cache.go` | 11 處 log 呼叫遷移 |
| `internal/chat/chat.go` | log 遷移 + 新增 ~15 處通訊 debug log（連線/斷線/重連/送事件/收訊息/channel 保護） |
| `internal/player/player.go` | 26 處 log 呼叫遷移 + 移除 Logger 別名 |
| `internal/probe/probe.go` | 1 處 log 呼叫遷移 |
| `go.mod` / `go.sum` | 移除 `go-logging` 依賴 |
| `AGENTS.md` | 更新 logger 相關說明 |

## 風險

1. **`log/slog` 需 Go 1.21+**：go.mod 目前為 `go 1.26.0`，完全相容
2. **`go-logging` 的 `Error()` 接受 `interface{}`**：slog 需要明確的 msg + key-value 參數，每個呼叫點都需要調整語法
3. **`player.go:411` 的 `Logger.Debug(duration)`**：傳入的是 float64 非字串，需改為 `slog.Debug("duration", "duration", duration)`
4. **`chat.go:61` 的 `logger.Logger.Error(r)`**：`r` 是 `recover()` 回傳的 `any`，需用 `fmt.Sprintf("%v", r)` 格式化為字串
5. **JSON 格式在開發時可讀性較低**：使用者已選擇 JSON 格式，適合生產環境日誌收集
6. **`player.go` 的 `Logger` 全域別名被其他檔案引用**：需確認無外部依賴後移除
7. **Socket.IO client event callback 簽名**：`connect_error` 的 callback 接收 `*client.ConnectError`，`disconnect` 接收 `string`（reason），`reconnect` / `reconnect_attempt` 接收 `int`（attempt number）。需確認 `zishang520/socket.io` v3 的 callback 簽名，用 `...any` 搭配型別斷言處理
8. **`SendEvent` debug log 頻率**：`TrackPlaying` 每秒呼叫 `FirePlaying` → `SendEvent`，debug log 會每秒產生一筆。這是可接受的（仅在 `LOG_LEVEL=DEBUG` 時輸出），但需注意 JSON handler 的效能

## 驗證步驟

1. 設定 `LOG_LEVEL=DEBUG`，執行 `make build`，確認編譯通過
2. 執行 `go vet ./...` 確認無警告
3. 執行 `make test` 確認測試通過
4. 手動驗證 JSON 輸出格式正確
5. 驗證 `LOG_LEVEL` 切換效果（DEBUG 應輸出更多日誌）
6. 驗證 chat 通訊 log：設定 `LOG_LEVEL=DEBUG`，觀察連線/斷線/重連/送事件/收訊息是否有完整的 JSON log 輸出
