package cache

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileCache(t *testing.T) {
	cache := NewFileCache("./test-cache")
	if cache == nil {
		t.Fatal("Expected FileCache instance, got nil")
	}

	if cache.basePath != "./test-cache" {
		t.Errorf("Expected basePath to be './test-cache', got '%s'", cache.basePath)
	}
}

func TestGenerateKey(t *testing.T) {
	cache := NewFileCache("./test-cache")
	key := cache.generateKey("test-url")

	// MD5 hash should be 32 characters
	if len(key) != 32 {
		t.Errorf("Expected key length to be 32, got %d", len(key))
	}

	// Calculate the expected MD5 hash for "test-url"
	expectedKey := fmt.Sprintf("%x", md5.Sum([]byte("test-url")))
	if key != expectedKey {
		t.Errorf("Expected key to be '%s', got '%s'", expectedKey, key)
	}
}

func TestFlush(t *testing.T) {
	// Create a temporary cache directory
	tempDir := "./test-cache-flush"
	cache := NewFileCache(tempDir)

	// Create some test files
	dataDir := filepath.Join(tempDir, "data")
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		t.Fatal("Failed to create test directory:", err)
	}

	testFile := filepath.Join(dataDir, "testfile")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatal("Failed to create test file:", err)
	}

	// Verify file exists before flush
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Test file should exist before flush")
	}

	// Flush the cache
	err = cache.Flush()
	if err != nil {
		t.Error("Flush should not return error:", err)
	}

	// Verify directory is removed
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Error("Data directory should be removed after flush")
	}

	// Clean up
	os.RemoveAll(tempDir)
}
