# mmfm-playback-go

為 `MMFM` 提供后台播放器。

此項目是對 `ffplay` 封裝執行播放任務，並通過 `websocket`([socket.io](https://socket.io)) 與 `MMFM` 通訊同步播放器狀態。

用於嵌入式系統或者無桌面 `linux` 系統的音頻播放。

`ffmpeg` 的可執行文件可以在[這裡](https://ffbinaries.com/downloads)下載，
不過建議 linux 系統直接 `yum` 及 `apt` 安裝。

請不要嘗試在 `docker` 運行，linux 下的聲音驅動是個深坑。

## 項目依賴

- [socket.io](https://socket.io)，成熟的 `websocket` 多端通訊方案。
- [ffmpeg 4.x](https://www.ffmpeg.org/)，成熟的流媒體播放及編碼/解碼方案。

## 項目結構

```
mmfm-playback-go/
├── cmd/
│   └── mmfm-playback/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── player/
│   │   ├── player.go
│   │   └── mplayer.go
│   ├── cache/
│   │   └── cache.go
│   ├── chat/
│   │   └── chat.go
│   ├── probe/
│   │   └── probe.go
│   └── logger/
│       └── logger.go
├── pkg/
│   └── types/
│       └── types.go
├── configs/
│   └── config.json
├── docs/
├── build/
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## 配置管理

項目支持多種配置格式：

### JSON 配置文件
默認為 `config.json`，也可以通過 `-c` 參數指定配置文件路徑：

```bash
./mmfm-playback-go -c ./myconfig.json
```

### 環境變量配置
支持以下環境變量：

- `FFPLAY_PATH` - ffplay 執行文件位置
- `FFPROBE_PATH` - ffprobe 執行文件位置
- `MPLAYER_PATH` - mplayer 執行文件位置
- `WEBSOCKET_API` - MMFM WebSocket 通訊地址
- `WEB_API` - MMFM 獲取歌曲地址 API
- `CACHE_PATH` - 音頻文件緩存位置

環境變量的優先級高於配置文件中的值。

## 編譯項目

```bash
go build -o mmfm-playback-go ./cmd/mmfm-playback
```

## 執行

```bash
./mmfm-playback-go -c ./configs/config.json
```

## Docker

- 編譯 `image`
```shell
docker build -t mmfm-playback-go .
```

- 執行
```shell
docker run -d --env-file .env mmfm-playback-go
```

## 配置文件說明

```json
{
    "ffmpeg": {
        "ffplay": "...", 
        "ffprobe": "...",
        "mplayer": "..."
    },
    "ws": "ws://localhost:8888/io/?EIO=3&transport=websocket",
    "cache": "...",
    "web": "http://localhost:8888/song/get"
}
```

|key|說明|
|-|-|
|ffmpeg.ffplay|ffplay 執行文件位置，linux下使用 which ffplay獲取|
|ffmpeg.ffprobe|ffprobe 執行文件位置，linux下使用 which ffprobe獲取|
|ffmpeg.mplayer|mplayer 執行文件位置|
|ws|`mmfm` websocket 通訊地址|
|cache|音頻文件緩存位置，建議使用系統臨時目錄，重新即燒毀|
|web|`mmfm` 獲取歌曲地址api|