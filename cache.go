package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type FileCache struct {
	basePath string
}

func NewFileCache(basePath string) *FileCache {
	return &FileCache{
		basePath: basePath,
	}
}

func (this *FileCache) Flush() error {
	return os.RemoveAll(filepath.Join(this.basePath, "data"))
}

func (this *FileCache) Cache(key string) (string) {
	hash := md5.Sum([]byte(key))
	hashKey := fmt.Sprintf("%x", hash)

	path := filepath.Join(this.basePath, "data", hashKey)
	if _, err := os.Stat(path); err == nil {
		log.Info("cache hint ", path)
		return path
	}

	go func() {
		log.Debug("begin cache music file")

		path := filepath.Join(this.basePath, "data", hashKey)
		dir := filepath.Dir(path)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, 0777)
		}

		resp, err := http.Get(key)
		if err != nil {
			log.Error(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Error("url return:", resp.StatusCode)
			return
		}

		out, err := os.Create(path)
		if err != nil {
			log.Error(err)
			return
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			if err != nil {
				log.Error(err)
				return
			}
		}

		log.Debug("cache music file:", path)
	}()

	return key
}