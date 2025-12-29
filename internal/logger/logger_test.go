package logger

import (
	"testing"
)

func TestLoggerInitialization(t *testing.T) {
	// Test that the logger is initialized properly
	if Logger == nil {
		t.Fatal("Expected Logger to be initialized, got nil")
	}

	// We can't easily test the actual logging functionality without capturing output
	// But we can at least verify the logger object exists
	t.Log("Logger initialized successfully")
}
