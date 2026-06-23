package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sumitlovanshi/url_shortener/internal/shortener"
)

func TestBboltStorePersistsLinksAcrossReopen(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "links.db")
	store := openTestStore(t, path)

	link := testLink("abc12345", "https://example.com/")
	if err := store.Save(context.Background(), link); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if _, err := store.IncrementClick(context.Background(), link.Code); err != nil {
		t.Fatalf("IncrementClick() error = %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	reopened := openTestStore(t, path)
	t.Cleanup(func() {
		_ = reopened.Close()
	})

	got, err := reopened.GetByCode(context.Background(), link.Code)
	if err != nil {
		t.Fatalf("GetByCode() error = %v", err)
	}
	if got.CanonicalURL != link.CanonicalURL {
		t.Fatalf("CanonicalURL = %q, want %q", got.CanonicalURL, link.CanonicalURL)
	}
	if got.ClickCount != 1 {
		t.Fatalf("ClickCount = %d, want 1", got.ClickCount)
	}

	code, err := reopened.FindCodeByCanonicalURL(context.Background(), link.CanonicalURL)
	if err != nil {
		t.Fatalf("FindCodeByCanonicalURL() error = %v", err)
	}
	if code != link.Code {
		t.Fatalf("code = %q, want %q", code, link.Code)
	}
}

func TestBboltStoreRejectsDuplicateCode(t *testing.T) {
	t.Parallel()

	store := openTestStore(t, filepath.Join(t.TempDir(), "links.db"))
	t.Cleanup(func() {
		_ = store.Close()
	})

	link := testLink("abc12345", "https://example.com/")
	if err := store.Save(context.Background(), link); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	duplicate := testLink("abc12345", "https://other.example/")
	err := store.Save(context.Background(), duplicate)
	if !errors.Is(err, shortener.ErrCodeExists) {
		t.Fatalf("Save() error = %v, want %v", err, shortener.ErrCodeExists)
	}
}

func TestBboltStoreRejectsDuplicateGeneratedURL(t *testing.T) {
	t.Parallel()

	store := openTestStore(t, filepath.Join(t.TempDir(), "links.db"))
	t.Cleanup(func() {
		_ = store.Close()
	})

	link := testLink("abc12345", "https://example.com/")
	if err := store.Save(context.Background(), link); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	duplicateURL := testLink("xyz98765", "https://example.com/")
	err := store.Save(context.Background(), duplicateURL)
	if !errors.Is(err, shortener.ErrURLExists) {
		t.Fatalf("Save() error = %v, want %v", err, shortener.ErrURLExists)
	}
}

func TestBboltStoreAllowsCustomAliasForSameURL(t *testing.T) {
	t.Parallel()

	store := openTestStore(t, filepath.Join(t.TempDir(), "links.db"))
	t.Cleanup(func() {
		_ = store.Close()
	})

	link := testLink("abc12345", "https://example.com/")
	if err := store.Save(context.Background(), link); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	alias := testLink("custom-alias", "https://example.com/")
	alias.CustomAlias = true
	if err := store.Save(context.Background(), alias); err != nil {
		t.Fatalf("Save(custom alias) error = %v", err)
	}
}

func openTestStore(t *testing.T, path string) *BboltStore {
	t.Helper()

	store, err := OpenBbolt(path)
	if err != nil {
		t.Fatalf("OpenBbolt() error = %v", err)
	}
	return store
}

func testLink(code, canonicalURL string) shortener.Link {
	return shortener.Link{
		Code:         code,
		OriginalURL:  canonicalURL,
		CanonicalURL: canonicalURL,
		CreatedAt:    time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC),
	}
}
