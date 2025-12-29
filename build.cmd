@echo off
echo Building mmfm-playback-go...

go build -mod=readonly -o bin/mmfm-playback-go.exe ./cmd/mmfm-playback

if %ERRORLEVEL% EQU 0 (
    echo Build successful! Binary located at bin/mmfm-playback-go.exe
) else (
    echo Build failed!
    exit /b %ERRORLEVEL%
)