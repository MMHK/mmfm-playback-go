# 配置管理系統

## 概述

mmfm-playback-go 項目實現了一個靈活的配置管理系統，支持多種配置來源和格式，包括 JSON 文件和環境變量。

## 配置結構

### 主要配置類型

#### PlaybackConfig
```go
type PlaybackConfig struct {
    FFMpegConf   *FFmpegConfig `json:"ffmpeg"`
    WebSocketAPI string        `json:"ws"`
    WebAPI       string        `json:"web"`
    CachePath    string        `json:"cache"`
    configFile   string
}
```

#### FFmpegConfig
```go
type FFmpegConfig struct {
    FFPlay  string `json:"ffplay"`
    FFProbe string `json:"ffprobe"`
    MPlayer string `json:"mplayer"`
}
```

## 配置來源優先級

配置系統按照以下優先級順序加載配置：

1. **環境變量** (最高優先級)
2. **JSON 配置文件** (默認優先級)

### 環境變量映射

| 環境變量 | 配置字段 | 說明 |
|----------|----------|------|
| FFPLAY_PATH | ffmpeg.ffplay | ffplay 執行文件路徑 |
| FFPROBE_PATH | ffmpeg.ffprobe | ffprobe 執行文件路徑 |
| MPLAYER_PATH | ffmpeg.mplayer | mplayer 執行文件路徑 |
| WEBSOCKET_API | ws | WebSocket API 地址 |
| WEB_API | web | Web API 地址 |
| CACHE_PATH | cache | 緩存目錄路徑 |
| WS_API | ws | (兼容) WebSocket API 地址 |
| WEB_API_URL | web | (兼容) Web API 地址 |
| CACHE_DIR | cache | (兼容) 緩存目錄路徑 |

## 配置加載流程

1. 嘗試從指定的 JSON 文件加載配置
2. 如果文件加載失敗，記錄警告並繼續
3. 從環境變量加載配置（覆蓋文件中的值）
4. 驗證必需的配置字段
5. 返回配置實例

## 使用示例

### JSON 配置文件

```json
{
    "ffmpeg": {
        "ffplay": "/usr/bin/ffplay", 
        "ffprobe": "/usr/bin/ffprobe",
        "mplayer": "/usr/bin/mplayer"
    },
    "ws": "ws://localhost:8888/io/?EIO=3&transport=websocket",
    "cache": "./cache",
    "web": "http://localhost:8888/song/get"
}
```

### 環境變量配置

```bash
export FFPLAY_PATH=/usr/local/bin/ffplay
export FFPROBE_PATH=/usr/local/bin/ffprobe
export MPLAYER_PATH=/usr/local/bin/mplayer
export WEBSOCKET_API=ws://production-server/ws
export WEB_API=http://production-server/api
export CACHE_PATH=/data/cache
```

### 代碼中使用配置

```go
conf, err := config.NewConfig("config.json")
if err != nil {
    logger.Logger.Error(err)
    return
}

// 使用配置
mp := player.NewMusicPlayer(conf)
```

## 配置驗證

配置系統會驗證以下必需字段：

- `ffmpeg.ffplay`: FFplay 執行文件路徑
- `ffmpeg.ffprobe`: FFprobe 執行文件路徑
- `ws`: WebSocket API 地址
- `web`: Web API 地址
- `cache`: 緩存目錄路徑

如果任何必需字段缺失，配置加載將返回錯誤。

## 配置保存

配置對象提供 `Save()` 方法，可以將當前配置保存到 JSON 文件：

```go
err := conf.Save()
if err != nil {
    logger.Logger.Error("Failed to save config:", err)
}
```

## 最佳實踐

### 開發環境
- 使用本地 JSON 配置文件進行開發
- 通過環境變量覆蓋特定設置

### 生產環境
- 使用環境變量進行配置
- 避免將敏感信息硬編碼到配置文件中
- 使用 Docker 時通過環境變量傳遞配置

### Docker 部署
- 使用 `.env` 文件管理環境變量
- 在 Dockerfile 中設置默認值
- 使用 volume 掛載配置文件（如果需要）

## 向後兼容性

配置系統支持舊版環境變量名稱，確保平滑遷移：

- `WS_API` 仍可作為 `ws` 的替代
- `WEB_API_URL` 仍可作為 `web` 的替代
- `CACHE_DIR` 仍可作為 `cache` 的替代

## 錯誤處理

配置加載可能遇到以下錯誤：

- 文件不存在或無法訪問
- JSON 格式錯誤
- 必需字段缺失
- 環境變量格式錯誤

所有錯誤都包含詳細的錯誤信息，便於調試。

## 擴展性

配置系統設計為可擴展：

- 可以輕鬆添加新的配置字段
- 支持新的配置來源（如遠程配置服務）
- 接口抽象允許替換不同的配置實現