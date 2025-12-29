# MMFM Playback Go 項目重構規劃

## 摘要

本文檔記錄了將 mmfm-playback-go 項目從單體結構重構為現代 Go 項目結構的完整規劃、過程和實施細節。

## 原始項目結構分析

### 問題點
- 所有代碼都在 main 包中
- 缺乏清晰的包結構
- 使用較舊的 Go 版本 (1.13)
- 功能混雜在一起（播放、緩存、配置、聊天等）
- 缺乏清晰的接口定義
- 測試覆蓋率低

### 原始文件結構
```
.
├── bin
├── Dockerfile
├── README.md
├── build.cmd
├── cache.go
├── chat.go
├── chat_test.go
├── conf.json
├── config.go
├── config.json
├── config_test.go
├── docker-compose-build.yml
├── docker-compose.yml
├── go.mod
├── go.sum
├── http.go
├── http_test.go
├── logger.go
├── main.go
├── player.go
├── player_test.go
├── testing.go
├── wrapper.go
└── wrapper_test.go
```

## 重構後的現代 Go 項目結構

```
mmfm-playback-go/
├── cmd/
│   └── mmfm-playback/
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
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
├── tests/
├── build/
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## 重構任務詳情

### 1. 包結構重構

#### internal/config
- 配置管理，支持 JSON 文件和環境變量
- 配置驗證和加載邏輯
- 配置保存功能

#### internal/player
- 播放器核心邏輯
- 播放、暫停、下一首等功能
- 播放狀態跟蹤

#### internal/cache
- 文件緩存系統
- 緩存清理和管理
- 文件下載和存儲

#### internal/chat
- WebSocket 通訊
- 消息處理和事件發送
- Socket.IO 集成

#### internal/probe
- 媒體文件分析
- FFprobe 集成
- 時長和格式檢測

#### internal/logger
- 統一日誌系統
- 日誌級別管理
- 格式化輸出

#### pkg/types
- 共享數據類型
- Song 結構體定義
- 通用接口定義

### 2. 配置管理系統

#### JSON 配置支持
- 從 JSON 文件加載配置
- 配置結構驗證
- 配置保存功能

#### 環境變量支持
- 環境變量覆蓋配置文件
- 靈活的部署選項
- Docker 容器友好

### 3. 接口抽象

定義了以下關鍵接口：

```go
type Player interface {
    Start() error
    Play(song *types.Song, second int) error
    Next()
    GetSongInPlayList(index int) (*types.Song, error)
}

type Cache interface {
    Cache(key string) string
    Clean(playlist []*types.Song) error
    Flush() error
}

type ChatClient interface {
    Listen() (<-chan *MessageArgs, error)
    SendEvent(eventName string, params *MessageArgs) error
    Close()
}
```

### 4. 測試改進

- 為各組件添加單元測試
- 配置加載測試
- 環境變量覆蓋測試
- 驗證測試

### 5. 构建和部署改進

#### Docker 支持
- 多階段構建
- 最小化鏡像
- 安全非 root 用戶

#### 構建腳本
- Windows 批處理文件
- Makefile 支持
- 跨平台構建

## 實施步驟

### 第一階段：基礎結構
1. 創建新的目錄結構
2. 遷移代碼到相應包中
3. 更新 go.mod 文件

### 第二階段：配置系統
1. 重構配置管理
2. 添加環境變量支持
3. 實現配置驗證

### 第三階段：核心功能
1. 重構播放器邏輯
2. 分離緩存功能
3. 重構聊天系統

### 第四階段：測試和驗證
1. 添加單元測試
2. 驗證構建過程
3. 更新文檔

## 驗證

項目已成功構建並通過了基本測試，驗證了：

- 構建系統正常工作
- 配置管理系統按預期工作
- 各組件能夠正確協同工作

## 部署指南

### 本地運行
```bash
go build -o mmfm-playback-go ./cmd/mmfm-playback
./mmfm-playback-go -c configs/config.json
```

### 使用環境變量
```bash
export WEBSOCKET_API=ws://your-server
export WEB_API=http://your-api
./mmfm-playback-go
```

### Docker 部署
```bash
docker build -t mmfm-playback-go .
docker run -d --env-file .env mmfm-playback-go
```

## 總結

這次重構成功地將一個單體 Go 項目轉換為現代化的、模塊化的結構，具有以下優勢：

1. **可維護性**：清晰的包結構使代碼更易於理解和維護
2. **可測試性**：接口抽象使單元測試更容易
3. **可擴展性**：模塊化設計支持將來的擴展
4. **靈活性**：支持多種配置方式，適應不同部署環境
5. **標準化**：遵循 Go 社區的項目結構最佳實踐