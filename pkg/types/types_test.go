package types

import "testing"

func TestSongGetURL(t *testing.T) {
	// Test when URL is set
	song := &Song{
		URL:  "http://example.com/song.mp3",
		Src:  "http://example.com/song-from-src.mp3",
		Name: "Test Song",
	}

	url := song.GetURL()
	if url != "http://example.com/song.mp3" {
		t.Errorf("Expected URL to be 'http://example.com/song.mp3', got '%s'", url)
	}

	// Test when URL is empty, should use Src
	song.URL = ""
	url = song.GetURL()
	if url != "http://example.com/song-from-src.mp3" {
		t.Errorf("Expected URL to be 'http://example.com/song-from-src.mp3', got '%s'", url)
	}

	// Test when both are empty
	song.URL = ""
	song.Src = ""
	url = song.GetURL()
	if url != "" {
		t.Errorf("Expected URL to be empty, got '%s'", url)
	}
}
