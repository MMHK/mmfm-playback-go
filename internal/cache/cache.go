package cache

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log/slog"
	"mmfm-playback-go/pkg/types"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Cache interface defines the caching functionality
type Cache interface {
	Cache(key string) string
	Clean(playlist []*types.Song) error
	Flush() error
}

// FileCache implements file-based caching
type FileCache struct {
	basePath string
}

// NewFileCache creates a new FileCache instance
func NewFileCache(basePath string) *FileCache {
	return &FileCache{
		basePath: basePath,
	}
}

// Flush removes all cached files
func (fc *FileCache) Flush() error {
	return os.RemoveAll(filepath.Join(fc.basePath, "data"))
}

// Clean removes cached files that are not in the playlist
func (fc *FileCache) Clean(playlist []*types.Song) error {
	allCaches, err := filepath.Glob(filepath.Join(fc.basePath, "data", "*"))
	if err != nil {
		slog.Error("clean cache failed", "error", err)
		return err
	}

	mapHash := []string{}
	for _, song := range playlist {
		mapHash = append(mapHash, fc.generateKey(song.GetURL()))
	}

	for _, path := range allCaches {
		cache := filepath.Base(path)

		for _, hash := range mapHash {
			if strings.EqualFold(hash, cache) {
				goto skip
			}
		}
		os.Remove(path)
	skip:
	}

	return nil
}

// generateKey generates a unique key for a URL
func (fc *FileCache) generateKey(key string) string {
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%x", hash)
}

// Cache caches a file from a URL if not already cached
func (fc *FileCache) Cache(key string) string {
	hashKey := fc.generateKey(key)

	path := filepath.Join(fc.basePath, "data", hashKey)
	if _, err := os.Stat(path); err == nil {
		slog.Info("cache hit", "path", path)
		return path
	}

	go func() {
		slog.Debug("begin cache music file")

		path := filepath.Join(fc.basePath, "data", hashKey)
		dir := filepath.Dir(path)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, 0777)
		}

		isHTTP := strings.HasPrefix(key, "http")

		if isHTTP {
			resp, err := http.Get(key)
			if err != nil {
				slog.Error("http get failed", "url", key, "error", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				slog.Error("url return non-200", "url", key, "status_code", resp.StatusCode)
				return
			}

			out, err := os.Create(path)
			if err != nil {
				slog.Error("create cache file failed", "path", path, "error", err)
				return
			}
			defer out.Close()
			// Write the body to file
			_, err = io.Copy(out, resp.Body)
			if err != nil {
				slog.Error("copy body to cache failed", "path", path, "error", err)
				return
			}
		} else {
			content, err := ioutil.ReadFile(key)
			if err != nil {
				slog.Error("read local file failed", "path", key, "error", err)
				return
			}

			out, err := os.Create(path)
			if err != nil {
				slog.Error("create cache file failed", "path", path, "error", err)
				return
			}
			defer out.Close()
			_, err = out.Write(content)
			if err != nil {
				slog.Error("write cache file failed", "path", path, "error", err)
			}
		}

		slog.Debug("cache music file done", "path", path)
	}()

	return key
}
