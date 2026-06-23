package shortener

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const maxGenerationAttempts = 10

type Service struct {
	repo      Repository
	generator CodeGenerator
	now       func() time.Time
}

func NewService(repo Repository, generator CodeGenerator) *Service {
	return &Service{
		repo:      repo,
		generator: generator,
		now:       time.Now,
	}
}

func (s *Service) Shorten(ctx context.Context, input ShortenInput) (ShortenResult, error) {
	canonicalURL, err := NormalizeURL(input.URL)
	if err != nil {
		return ShortenResult{}, err
	}

	if input.Alias != nil {
		return s.shortenWithAlias(ctx, canonicalURL, *input.Alias)
	}

	code, err := s.repo.FindCodeByCanonicalURL(ctx, canonicalURL)
	if err == nil {
		link, getErr := s.repo.GetByCode(ctx, code)
		return ShortenResult{Link: link, Reused: true}, getErr
	}
	if !errors.Is(err, ErrNotFound) {
		return ShortenResult{}, err
	}

	for attempt := 0; attempt < maxGenerationAttempts; attempt++ {
		code, err := s.generator.Generate(ctx)
		if err != nil {
			return ShortenResult{}, err
		}

		link := Link{
			Code:         code,
			OriginalURL:  canonicalURL,
			CanonicalURL: canonicalURL,
			CreatedAt:    s.now().UTC(),
			CustomAlias:  false,
		}

		if err := s.repo.Save(ctx, link); err == nil {
			return ShortenResult{Link: link}, nil
		} else if errors.Is(err, ErrCodeExists) {
			continue
		} else if errors.Is(err, ErrURLExists) {
			return s.reuseExisting(ctx, canonicalURL)
		} else {
			return ShortenResult{}, err
		}
	}

	return ShortenResult{}, fmt.Errorf("generate unique code: %w", ErrCodeExists)
}

func (s *Service) RecordRedirect(ctx context.Context, code string) (Link, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return Link{}, ErrNotFound
	}
	return s.repo.IncrementClick(ctx, code)
}

func (s *Service) GetStats(ctx context.Context, code string) (Link, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return Link{}, ErrNotFound
	}
	return s.repo.GetByCode(ctx, code)
}

func (s *Service) shortenWithAlias(ctx context.Context, canonicalURL, alias string) (ShortenResult, error) {
	alias = strings.TrimSpace(alias)
	if err := ValidateAlias(alias); err != nil {
		return ShortenResult{}, err
	}

	link := Link{
		Code:         alias,
		OriginalURL:  canonicalURL,
		CanonicalURL: canonicalURL,
		CreatedAt:    s.now().UTC(),
		CustomAlias:  true,
	}

	if err := s.repo.Save(ctx, link); err != nil {
		if errors.Is(err, ErrCodeExists) {
			return ShortenResult{}, ErrAliasConflict
		}
		return ShortenResult{}, err
	}

	return ShortenResult{Link: link}, nil
}

func (s *Service) reuseExisting(ctx context.Context, canonicalURL string) (ShortenResult, error) {
	code, err := s.repo.FindCodeByCanonicalURL(ctx, canonicalURL)
	if err != nil {
		return ShortenResult{}, err
	}
	link, err := s.repo.GetByCode(ctx, code)
	return ShortenResult{Link: link, Reused: true}, err
}
