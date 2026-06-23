package shortener

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestShortenReusesExistingGeneratedURL(t *testing.T) {
	t.Parallel()

	repo := newMemoryRepo()
	service := NewService(repo, &sequenceGenerator{codes: []string{"abc12345", "zzz99999"}})
	service.now = fixedNow

	first, err := service.Shorten(context.Background(), ShortenInput{URL: "https://Example.com"})
	if err != nil {
		t.Fatalf("first Shorten() error = %v", err)
	}

	second, err := service.Shorten(context.Background(), ShortenInput{URL: "https://example.com/"})
	if err != nil {
		t.Fatalf("second Shorten() error = %v", err)
	}

	if first.Link.Code != second.Link.Code {
		t.Fatalf("second code = %q, want reused code %q", second.Link.Code, first.Link.Code)
	}
	if !second.Reused {
		t.Fatal("second Shorten() Reused = false, want true")
	}
}

func TestShortenWithCustomAlias(t *testing.T) {
	t.Parallel()

	alias := "product_launch"
	repo := newMemoryRepo()
	service := NewService(repo, &sequenceGenerator{codes: []string{"unused"}})
	service.now = fixedNow

	result, err := service.Shorten(context.Background(), ShortenInput{
		URL:   "https://example.com/launch",
		Alias: &alias,
	})
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}

	if result.Link.Code != alias {
		t.Fatalf("code = %q, want %q", result.Link.Code, alias)
	}
	if !result.Link.CustomAlias {
		t.Fatal("CustomAlias = false, want true")
	}
	if result.Reused {
		t.Fatal("Reused = true, want false")
	}
}

func TestShortenRejectsDuplicateAlias(t *testing.T) {
	t.Parallel()

	alias := "same-alias"
	repo := newMemoryRepo()
	service := NewService(repo, &sequenceGenerator{codes: []string{"unused"}})
	service.now = fixedNow

	_, err := service.Shorten(context.Background(), ShortenInput{URL: "https://example.com/one", Alias: &alias})
	if err != nil {
		t.Fatalf("first Shorten() error = %v", err)
	}

	_, err = service.Shorten(context.Background(), ShortenInput{URL: "https://example.com/two", Alias: &alias})
	if !errors.Is(err, ErrAliasConflict) {
		t.Fatalf("second Shorten() error = %v, want %v", err, ErrAliasConflict)
	}
}

func TestShortenRetriesCodeCollision(t *testing.T) {
	t.Parallel()

	repo := newMemoryRepo()
	err := repo.Save(context.Background(), Link{
		Code:         "taken",
		OriginalURL:  "https://existing.example/",
		CanonicalURL: "https://existing.example/",
		CreatedAt:    fixedNow(),
	})
	if err != nil {
		t.Fatalf("preload Save() error = %v", err)
	}

	service := NewService(repo, &sequenceGenerator{codes: []string{"taken", "free"}})
	service.now = fixedNow

	result, err := service.Shorten(context.Background(), ShortenInput{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}
	if result.Link.Code != "free" {
		t.Fatalf("code = %q, want free", result.Link.Code)
	}
}

func TestRecordRedirectIncrementsClickCount(t *testing.T) {
	t.Parallel()

	repo := newMemoryRepo()
	service := NewService(repo, &sequenceGenerator{codes: []string{"abc12345"}})
	service.now = fixedNow

	result, err := service.Shorten(context.Background(), ShortenInput{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}

	link, err := service.RecordRedirect(context.Background(), result.Link.Code)
	if err != nil {
		t.Fatalf("RecordRedirect() error = %v", err)
	}
	if link.ClickCount != 1 {
		t.Fatalf("ClickCount = %d, want 1", link.ClickCount)
	}
}

type sequenceGenerator struct {
	codes []string
	next  int
}

func (g *sequenceGenerator) Generate(context.Context) (string, error) {
	if g.next >= len(g.codes) {
		return "", errors.New("no codes left")
	}
	code := g.codes[g.next]
	g.next++
	return code, nil
}

type memoryRepo struct {
	byCode map[string]Link
	byURL  map[string]string
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		byCode: make(map[string]Link),
		byURL:  make(map[string]string),
	}
}

func (r *memoryRepo) Save(_ context.Context, link Link) error {
	if _, exists := r.byCode[link.Code]; exists {
		return ErrCodeExists
	}
	if !link.CustomAlias {
		if _, exists := r.byURL[link.CanonicalURL]; exists {
			return ErrURLExists
		}
		r.byURL[link.CanonicalURL] = link.Code
	}
	r.byCode[link.Code] = link
	return nil
}

func (r *memoryRepo) FindCodeByCanonicalURL(_ context.Context, canonicalURL string) (string, error) {
	code, exists := r.byURL[canonicalURL]
	if !exists {
		return "", ErrNotFound
	}
	return code, nil
}

func (r *memoryRepo) GetByCode(_ context.Context, code string) (Link, error) {
	link, exists := r.byCode[code]
	if !exists {
		return Link{}, ErrNotFound
	}
	return link, nil
}

func (r *memoryRepo) IncrementClick(_ context.Context, code string) (Link, error) {
	link, exists := r.byCode[code]
	if !exists {
		return Link{}, ErrNotFound
	}
	link.ClickCount++
	r.byCode[code] = link
	return link, nil
}

func (r *memoryRepo) Close() error {
	return nil
}

func fixedNow() time.Time {
	return time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)
}
