package shortener

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidURL    = errors.New("invalid url")
	ErrInvalidAlias  = errors.New("invalid alias")
	ErrNotFound      = errors.New("link not found")
	ErrCodeExists    = errors.New("code already exists")
	ErrURLExists     = errors.New("url already exists")
	ErrAliasConflict = errors.New("alias already exists")
)

type Link struct {
	Code         string    `json:"code"`
	OriginalURL  string    `json:"url"`
	CanonicalURL string    `json:"canonical_url"`
	CreatedAt    time.Time `json:"created_at"`
	ClickCount   uint64    `json:"click_count"`
	CustomAlias  bool      `json:"custom_alias"`
}

type ShortenInput struct {
	URL   string
	Alias *string
}

type ShortenResult struct {
	Link   Link
	Reused bool
}

type Repository interface {
	Save(ctx context.Context, link Link) error
	FindCodeByCanonicalURL(ctx context.Context, canonicalURL string) (string, error)
	GetByCode(ctx context.Context, code string) (Link, error)
	IncrementClick(ctx context.Context, code string) (Link, error)
	Close() error
}
