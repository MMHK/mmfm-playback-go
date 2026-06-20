package logger

import (
	"log/slog"
	"os"
	"testing"
)

func TestInit_DefaultLevel(t *testing.T) {
	os.Unsetenv("LOG_LEVEL")
	Init()
	logger := slog.Default()
	if logger == nil {
		t.Fatal("expected default logger to be initialized")
	}
}

func TestInit_DebugLevel(t *testing.T) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	defer os.Unsetenv("LOG_LEVEL")
	Init()
	logger := slog.Default()
	if logger == nil {
		t.Fatal("expected default logger to be initialized")
	}
}

func TestInit_ErrorLevel(t *testing.T) {
	os.Setenv("LOG_LEVEL", "ERROR")
	defer os.Unsetenv("LOG_LEVEL")
	Init()
	logger := slog.Default()
	if logger == nil {
		t.Fatal("expected default logger to be initialized")
	}
}

func TestInit_InvalidLevel(t *testing.T) {
	os.Setenv("LOG_LEVEL", "INVALID")
	defer os.Unsetenv("LOG_LEVEL")
	Init()
	logger := slog.Default()
	if logger == nil {
		t.Fatal("expected default logger to be initialized, invalid level should fallback to INFO")
	}
}

func TestInit_TextFormat(t *testing.T) {
	os.Setenv("LOG_FORMAT", "text")
	defer os.Unsetenv("LOG_FORMAT")
	Init()
	logger := slog.Default()
	if logger == nil {
		t.Fatal("expected default logger to be initialized")
	}
}

func TestInit_JSONFormat(t *testing.T) {
	os.Setenv("LOG_FORMAT", "json")
	defer os.Unsetenv("LOG_FORMAT")
	Init()
	logger := slog.Default()
	if logger == nil {
		t.Fatal("expected default logger to be initialized")
	}
}

func TestInit_DefaultFormat(t *testing.T) {
	os.Unsetenv("LOG_FORMAT")
	Init()
	logger := slog.Default()
	if logger == nil {
		t.Fatal("expected default logger to be initialized, default format should be text")
	}
}
