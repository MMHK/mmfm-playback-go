# Socket.IO SDK 升級計畫：v2 -> v4

## 目標

將 Go playback 專案的 Socket.IO 客戶端從 `graarh/golang-socketio`（Socket.IO v1/v2, EIO=3）升級為 `zishang520/socket.io/clients/socket/v3`（Socket.IO v4+, EIO=4），使其能正確連線到 MMFM 後端（Socket.IO v4.8.3, Engine.IO v6.6.9）。

## 現狀分析

| 項目 | 目前 | 目標 |
|------|------|------|
| Go 版本 | `go 1.21` | `go 1.26` |
| Go SDK | `graarh/golang-socketio` (2017) | `zishang520/socket.io/clients/socket/v3` v3.0.4 |
| Engine.IO 協定 | EIO=3 | EIO=4 |
| Socket.IO 協定 | v1/v2 | v4+ |
| MMFM 後端 | Socket.IO v4.8.3, path `/io` | 不變 |
| 自動重連 | 手動（遞迴呼叫 `Listen()`） | SDK 內建 `SetReconnection(true)` |
| WS URL 格式 | `ws://host:port/io/?EIO=3&transport=websocket` | `http://host:port` |
| 伺服器 path | 硬編碼在 URL | SDK options `SetPath("/io")` |

### 兼容性調查結論

- MMFM 後端使用 Socket.IO v4.8.3（Engine.IO v6.6.9），`path: "/io"`，未設定 `allowEIO3`
- `graarh/golang-socketio` 硬編碼 `EIO=3`，使用 Engine.IO v3 協定 — **無法連線**
- `zishang520/socket.io/clients/socket/v3` 需要 Go >= 1.26.0 — 專案需從 1.21 升級
- Go 1.26.0 已安裝於 `C:\Users\Sam\sdk\go1.26.0\bin\go.exe`

## 任務清單

### 0. Go 版本升級
- [x] `go.mod` 將 `go 1.21` 改為 `go 1.26.0`
- [x] 驗證 `go build ./...` 在 Go 1.26 下編譯通過（確認無 breaking changes）

### 1. 依賴替換
- [x] `go.mod` 移除 `github.com/graarh/golang-socketio`
- [x] `go.mod` 移除間接依賴 `github.com/gorilla/websocket v1.4.0`（保留 v1.5.3 作為 zishang520 的間接依賴）
- [x] 新增 `github.com/zishang520/socket.io/clients/socket/v3` v3.0.4
- [x] 執行 `go mod tidy` 清理

### 2. 核心重構：`internal/chat/chat.go`
- [x] 替換 import：`gosocketio` -> `socketio`（zishang520 客戶端）
- [x] 移除 import：`gosocketio/transport`
- [x] 新增 import：`zishang520/socket.io/clients/engine/v3/transports`、`zishang520/socket.io/v3/pkg/types`
- [x] `ChatClient` struct：`*gosocketio.Client` -> `*socketio.Socket`
- [x] `Connect()` 方法：
  - 使用 `socketio.DefaultOptions()` 建立 options
  - `opts.SetPath("/io")` 設定伺服器 path
  - `opts.SetTransports(siotypes.NewSet(transports.WebSocket))` 限定 WebSocket
  - `opts.SetReconnection(true)` 啟用自動重連
  - `opts.SetReconnectionAttempts(999999)` 無限重連
  - 使用 `socketio.Connect(url, opts)` 連線
- [x] `OnConnection` / `OnDisconnection` 常數 -> 字串 `"connect"` / `"disconnect"`
- [x] 事件回呼簽名：`func(h *gosocketio.Channel, data string)` -> `func(data ...any)`
- [x] `msg` 事件處理：從 `...any` 中提取字串參數（含 type switch fallback），轉為 JSON 後餵給 `ParseMessageArgs`
- [x] `Emit`：適配新 API `socket.Emit(event, args...)`
- [x] `Close()`：改用 `socket.Disconnect()`
- [x] 移除手動重連邏輯（遞迴 `Listen()`），改用 SDK 內建 `SetReconnection(true)`
- [x] 新增 `disconnect` 事件處理：記錄 log

### 3. 設定檔 URL 格式更新
所有 WS URL 從 `ws://host:port/io/?EIO=3&transport=websocket` 簡化為 `http://host:port`

- [x] `configs/config.json` line 7
- [x] `config.json`（模板）line 7
- [x] `conf.json` line 7
- [x] `.env.example` line 9, 16
- [x] `docker-compose.yml` line 14

### 4. 測試驗證
- [x] `internal/chat/chat_test.go`：確認現有測試不需改動（未直接引用 SDK），3/3 通過
- [x] `go vet ./...` 通過（僅有 pre-existing warnings：config_test.go:62, player.go:320）
- [x] `go build ./...` 成功
- [x] `go test ./internal/chat/... ./pkg/...` 通過
- [ ] failed `make test`：`internal/config/config_test.go:62` pre-existing build error（`no new variables on left side of :=`），非本次變更引入

### 5. 整合驗證（需 MMFM 後端運行）
- [ ] Go client 成功連線到 MMFM Socket.IO v4 伺服器
- [ ] `msg` 事件收發正常
- [ ] 自動重連在斷線後正常觸發
- [ ] 播放狀態同步（`player.playing`, `player.pause`）正常

## 受影響檔案

| 檔案 | 變更類型 |
|------|---------|
| `go.mod` | Go 版本 + 依賴替換 |
| `go.sum` | 自動更新 |
| `internal/chat/chat.go` | 核心重構 |
| `configs/config.json` | URL 格式 |
| `config.json` | URL 格式 |
| `conf.json` | URL 格式 |
| `.env.example` | URL 格式 |
| `docker-compose.yml` | URL 格式 |

**不受影響：** `internal/player/player.go`（透過 `ChatClient` 抽象層隔離）、`internal/config/config.go`、`internal/cache/`、`internal/probe/`

## API 對照表

| 舊 (graarh) | 新 (zishang520) |
|-------------|----------------|
| `gosocketio.Dial(url, transport)` | `socket.Connect(url, opts)` |
| `transport.GetDefaultWebsocketTransport()` | `opts.SetTransports(types.NewSet(transports.WebSocket))` |
| `gosocketio.OnConnection` | `"connect"` |
| `gosocketio.OnDisconnection` | `"disconnect"` |
| `client.On(event, func(h *Channel))` | `socket.On(event, func(...any))` |
| `client.On(event, func(h *Channel, data string))` | `socket.On(event, func(data ...any))` |
| `client.Emit(event, data)` | `socket.Emit(event, data)` |
| `client.Close()` | `socket.Close()` |
| 手動遞迴 `Listen()` 重連 | `opts.SetReconnection(true)` |
| URL 含 `/io/?EIO=3&transport=websocket` | `opts.SetPath("/io")` + 純 host URL |

## 風險

1. **依賴體積** — zishang520/socket.io 是 monorepo，引入 engine.io、parser、quic-go、webtransport-go 等子模組，二進制體積會顯著增加
2. **Go 版本升級** — 從 1.21 跳到 1.26，需確認專案現有程式碼無 compatibility issue
3. **事件資料格式** — `msg` 事件的資料在 v4 中以 `...any` 形式接收，需確認 JSON 字串的解析邏輯是否相容
4. **`msg` 事件資料型別** — v2 回呼直接收到 `string`，v4 收到 `...any`；若 Socket.IO parser 將資料解為 `map[string]any` 而非字串，需調整 `ParseMessageArgs` 的處理方式

## 驗證步驟

1. `go mod tidy` 確認依賴解析成功
2. `go build ./...` 確認編譯通過
3. `go vet ./...` 確認無靜態分析問題
4. `make test` 確認現有測試通過
5. 啟動 MMFM 後端 + Go playback，確認 Socket.IO 連線建立
6. 測試播放/暫停/切歌等指令同步
