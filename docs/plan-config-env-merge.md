# 規劃：Config 合併 Env 流程

## 目標

在非測試流程中，實現 config 的 env 合併機制。優先級為：`.env` > `os env` > `JSON config`。

## 任務

- [x] 1. 在 `cmd/mmfm-playback/main.go` 加載 `.env` 檔案（若存在）
- [x] 2. 檢查 `internal/config/config.go` 的合併邏輯（確認無需修改）
- [x] 3. 確認 `loadFromEnv()` 在 `.env` 載入後執行
- [x] 4. 更新 `AGENTS.md` 說明新的配置優先級
- [x] 5. 驗證流程：JSON → os env → .env（審查確認邏輯正確）

## 合併策略

### 載入順序

1. **JSON config**：首先載入 `configs/config.json` 作為基礎值
2. **os env**：系統環境變數覆蓋 JSON 的值
3. **.env**：`.env` 檔案中的變數覆蓋 os env 和 JSON

### 環境變數映射規則

沿用現有 `loadFromEnv()` 的映射：

| 環境變數 | Config 欄位 |
|---------|------------|
| `FFPLAY_PATH` | `FFplayPath` |
| `FFPROBE_PATH` | `FFprobePath` |
| `MPLAYER_PATH` | `MplayerPath` |
| `WEBSOCKET_API` | `WebsocketAPI` |
| `WEB_API` | `WebAPI` |
| `CACHE_PATH` | `CachePath` |

### .env 檔案格式

```env
FFPLAY_PATH=/usr/bin/ffplay
WEBSOCKET_API=ws://localhost:8080
```

## 受影響的檔案

- `cmd/mmfm-playback/main.go`：新增 `.env` 載入邏輯
- `internal/config/config.go`：調整合併順序（若需要）
- `AGENTS.md`：更新配置說明

## 風險

- 現有的 `loadFromEnv()` 已經支援 os env，只需在 main 提前載入 `.env` 即可
- `godotenv.Load()` 不會覆蓋已存在的環境變數，需用 `godotenv.Overload()` 或手動處理優先級
- 需確保 `.env` 載入時機在 `NewConfig()` 之前

## 驗證步驟

1. 準備測試場景：
   - JSON config 設定 `CachePath = "/json/cache"`
   - os env 設定 `CACHE_PATH="/os/cache"`
   - `.env` 設定 `CACHE_PATH=/dotenv/cache`
2. 驗證最終 config 的 `CachePath` 為 `/dotenv/cache`
3. 移除 `.env`，驗證 `CachePath` 為 `/os/cache`
4. 清除 os env，驗證 `CachePath` 為 `/json/cache`
