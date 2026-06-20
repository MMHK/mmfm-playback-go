# AGENTS.md

## Project

Go 1.21 service that wraps `ffplay`/`mplayer`/`ffprobe` to play audio for an MMFM backend, communicating over Socket.IO (WebSocket). Targets embedded/headless Linux for audio playback.

## Commands

```bash
make build          # -> bin/mmfm-playback-go
make test           # go test -mod=readonly -v ./...
make test-coverage  # + coverage.html
build.cmd           # Windows build -> bin/mmfm-playback-go.exe
go vet ./...        # no linter configured; run this before committing
```

No lint, formatter, or CI config exists. There is no typecheck step beyond `go vet` / `go build`.

## Run

```bash
./bin/mmfm-playback-go -c ./configs/config.json
```

Env vars override JSON config: `FFPLAY_PATH`, `FFPROBE_PATH`, `MPLAYER_PATH`, `WEBSOCKET_API`, `WEB_API`, `CACHE_PATH`, `LOG_LEVEL`.

Config merge priority (highest to lowest): `.env` file > OS env vars > JSON config file.

## Architecture

```
cmd/mmfm-playback/main.go   entrypoint: -c flag -> config.NewConfig -> player.NewMusicPlayer -> Start()
internal/player/             core orchestrator: playlist, playback state, scheduled audio
internal/cache/              audio file download + local cache
internal/chat/               Socket.IO client to MMFM server (event sync)
internal/probe/              ffprobe wrapper for audio metadata
internal/config/             JSON file + env var loading, validation
internal/logger/             log/slog setup
pkg/types/                   shared Song struct
tests/testings.go            test helpers (LoadTestEnv, GetLocalPath) — not tests themselves
testdata/                    contains a test .mp3 fixture
```

Dependency flow: `player -> {cache, chat, config, probe, types}`.

## Conventions

- Project language is Traditional Chinese (comments, docs, commit messages).
- Module path: `mmfm-playback-go` (no domain prefix).
- Tests are `_test.go` alongside source. `tests/` dir holds helpers, not test cases.
- Config struct fields use `json` tags; new config keys must update both `PlaybackConfig` and `loadFromEnv()`.

## Workflow

1. **Understand first** — never start work immediately. Keep asking the user questions until requirements, scope, and acceptance criteria are fully aligned.
2. **Plan in `/docs/`** — create a planning MD at `docs/plan-<topic>.md` containing: goal, tasks (with `[ ]` checkboxes), affected files, risks, and verification steps.
3. **Confirm before acting** — present the plan and ask the user to confirm it is complete and correct before any implementation begins.
4. **Delegate via sub-agents** — the main agent MUST NOT write code directly. All implementation, analysis, and planning tasks are executed by sub-agents (use the Task tool).
5. **Test via sub-agent** — after implementation, dispatch a sub-agent to write or update test cases covering the new/changed code. Run `make test` and `go vet` to verify.
6. **Review completion via sub-agent** — dispatch a separate sub-agent to audit the plan against actual changes: verify every task is addressed, tests pass, no regressions, and the code matches the plan's intent. Report any gaps back to the main agent.
7. **Update plan status** — after tests pass and review confirms completeness, mark each task as `[x]` or `[ ] failed` in the planning MD and note any deviations.

## Known issues

- `internal/player/player.go:320` has unreachable code (flagged by `go vet`).
- `internal/config/config_test.go:62` has a `no new variables on left side of :=` warning.

## Gotchas

- Do not try to run this in Docker for real playback — audio device drivers do not pass through (noted in README).
- `config.json` at repo root uses `${VAR}` placeholders; it is a template, not a working config. Use `configs/config.json` or env vars.
- `go.mod` specifies `go 1.21` but uses only stdlib-compatible patterns; no generics-heavy deps.
- Socket.IO uses the `zishang520/socket.io/clients/socket/v3` library (Socket.IO v4+, Engine.IO v4).
