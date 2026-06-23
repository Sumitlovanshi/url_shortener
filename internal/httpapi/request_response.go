package httpapi

import "time"

type shortenRequest struct {
	URL   string  `json:"url"`
	Alias *string `json:"alias,omitempty"`
}

type shortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
	URL      string `json:"url"`
	Reused   bool   `json:"reused"`
}

type statsResponse struct {
	Code        string    `json:"code"`
	URL         string    `json:"url"`
	ClickCount  uint64    `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
	CustomAlias bool      `json:"custom_alias"`
}

type errorResponse struct {
	Error string `json:"error"`
}
