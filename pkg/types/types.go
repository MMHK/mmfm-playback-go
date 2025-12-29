package types

// Song represents a music track
type Song struct {
	Cover    string  `json:"cover"`
	URL      string  `json:"url"`
	Src      string  `json:"src"`
	Name     string  `json:"name"`
	Author   string  `json:"author"`
	Duration float64 `json:"duration"`
	Index    float64 `json:"index"`
}

// GetURL returns the URL of the song, preferring URL field over Src
func (s *Song) GetURL() string {
	url := s.URL
	if len(url) <= 0 {
		url = s.Src
	}
	return url
}
