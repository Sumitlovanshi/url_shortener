package shortener

import (
	"errors"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "lowercases scheme and host",
			raw:  "HTTP://Example.COM:80/a?b=c",
			want: "http://example.com/a?b=c",
		},
		{
			name: "adds root path",
			raw:  "https://example.com",
			want: "https://example.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeURL(tt.raw)
			if err != nil {
				t.Fatalf("NormalizeURL() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeURLRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []string{
		"",
		"example.com",
		"ftp://example.com/file",
		"https://",
		"https://user:pass@example.com",
		"https://example.com/with space",
	}

	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			_, err := NormalizeURL(raw)
			if !errors.Is(err, ErrInvalidURL) {
				t.Fatalf("NormalizeURL(%q) error = %v, want %v", raw, err, ErrInvalidURL)
			}
		})
	}
}

func TestValidateAlias(t *testing.T) {
	t.Parallel()

	valid := []string{"abc", "Campaign_2026", "team-link"}
	for _, alias := range valid {
		if err := ValidateAlias(alias); err != nil {
			t.Fatalf("ValidateAlias(%q) error = %v", alias, err)
		}
	}

	invalid := []string{"", "ab", "has space", "slash/path", "unicode-ह"}
	for _, alias := range invalid {
		if err := ValidateAlias(alias); !errors.Is(err, ErrInvalidAlias) {
			t.Fatalf("ValidateAlias(%q) error = %v, want %v", alias, err, ErrInvalidAlias)
		}
	}
}
