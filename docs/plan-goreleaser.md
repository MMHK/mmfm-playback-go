# 規劃：GoReleaser + GitHub Actions 自動_RELEASE

## 目標

設定 GoReleaser 自動編譯並發布 `mmfm-playback-go`，每次 push tag (`v*`) 時觸發 GitHub Actions，編譯多個 Linux 平台的二進制文件，並附帶 `.env.example` 和 `configs/config.json` 作為發布產物。

## 任務

- [x] 建立 `.goreleaser.yml` 設定檔
- [x] 建立 `.github/workflows/release.yml` GitHub Actions workflow
- [x] 確認 `.env.example` 存在於 repo（已存在）
- [x] 確認 `configs/config.json` 存在於 repo（已存在）
- [x] 修改 `cmd/mmfm-playback/main.go` 加入版本變數

## 詳細規劃

### 1. `.goreleaser.yml`

```yaml
version: 2

builds:
  - id: mmfm-playback-go
    main: ./cmd/mmfm-playback
    binary: mmfm-playback-go
    env:
      - CGO_ENABLED=0
    goos:
      - linux
    goarch:
      - amd64
      - arm64
      - arm
    goarm:
      - "7"
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - id: default
    name_template: >-
      {{ .ProjectName }}_
      {{- .Version }}_
      {{- .Os }}_
      {{- if eq .Arch "arm" }}armv7
      {{- else }}{{ .Arch }}{{ end }}
    files:
      - .env.example
      - configs/config.json
      - README.md
      - LICENSE

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^Merge"
```

**說明：**
- `CGO_ENABLED=0`：靜態編譯，適合嵌入式/無桌面 Linux
- `goarm: "7"` 對應 `linux/armv7` 平台
- archives 內含 `.env.example`、`configs/config.json`、`README.md`、`LICENSE`
- `ldflags` 注入版本資訊（需 main package 有對應變數）
- archive 命名範例：`mmfm-playback-go_v1.0.0_linux_amd64.tar.gz`

### 2. `.github/workflows/release.yml`

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**說明：**
- 觸發條件：push tag 匹配 `v*`（如 `v1.0.0`）
- `fetch-depth: 0`：GoReleaser 需要完整 git history 來生成 changelog
- `go-version-file`：直接從 `go.mod` 讀取 Go 版本
- `GITHUB_TOKEN`：GitHub Actions 內建 token，用於建立 Release

### 3. `main.go` 版本資訊變數（選用）

如要在 binary 中嵌入版本資訊，需在 `cmd/mmfm-playback/main.go` 加入：

```go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

然後在 `printServiceInfo` 中加入版本輸出。

## 受影響文件

| 文件 | 操作 |
|------|------|
| `.goreleaser.yml` | 新建 |
| `.github/workflows/release.yml` | 新建 |
| `cmd/mmfm-playback/main.go` | 修改（選用，加入版本變數） |

## 發布產物

每次 release 會產生：

| 檔案 | 說明 |
|------|------|
| `mmfm-playback-go_vX.Y.Z_linux_amd64.tar.gz` | x86_64 Linux |
| `mmfm-playback-go_vX.Y.Z_linux_arm64.tar.gz` | ARM64 Linux |
| `mmfm-playback-go_vX.Y.Z_linux_armv7.tar.gz` | ARMv7 Linux |
| `checksums.txt` | SHA256 校驗碼 |

每個 `.tar.gz` 會包含：`mmfm-playback-go` binary + `.env.example` + `configs/config.json` + `README.md` + `LICENSE`

## 風險

1. **Go 版本相容**：`go.mod` 指定 `go 1.26.0`，GitHub Actions 的 `setup-go` 需確認能正確安裝此版本
2. **無 CGO 依賴**：已設 `CGO_ENABLED=0`，不需要交叉編譯工具鏈
3. **tag 格式**：必須使用 `v` 開頭（如 `v1.0.0`），否则不会触发

## 驗證步驟

1. 建立 tag 前，先用 `goreleaser release --snapshot --clean` 本地測試
2. 確認產生的 tar.gz 包含所有預期檔案
3. 確認 binary 可在目標平台執行
4. Push tag 到 GitHub，確認 Actions 成功並產生 Release
