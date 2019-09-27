package main

import (
	"encoding/json"
	"net/http"
)


func LoadPlaylist(url string) ([]*Song, error) {
	resp, err := http.Get(url)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer resp.Body.Close()

	list := make([]*Song, 0)
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&list)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return list, nil
}