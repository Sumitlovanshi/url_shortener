package shortener

import "time"

type Link struct {
	Code         string    `json:"code"`
	OriginalURL  string    `json:"url"`
	CanonicalURL string    `json:"canonical_url"`
	CreatedAt    time.Time `json:"created_at"`
	ClickCount   uint64    `json:"click_count"`
	CustomAlias  bool      `json:"custom_alias"`
}
