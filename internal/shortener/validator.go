package shortener

import (
	"net"
	"net/url"
	"regexp"
	"strings"
)

var aliasPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,64}$`)

func NormalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.ContainsAny(raw, " \t\r\n") {
		return "", ErrInvalidURL
	}

	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() || parsed.Hostname() == "" {
		return "", ErrInvalidURL
	}
	if parsed.User != nil {
		return "", ErrInvalidURL
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", ErrInvalidURL
	}

	host := strings.ToLower(parsed.Hostname())
	port := parsed.Port()
	if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		port = ""
	}

	parsed.Scheme = scheme
	parsed.Host = host
	if port != "" {
		parsed.Host = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		parsed.Host = "[" + host + "]"
	}
	if parsed.Path == "" {
		parsed.Path = "/"
	}

	return parsed.String(), nil
}

func ValidateAlias(alias string) error {
	if !aliasPattern.MatchString(alias) {
		return ErrInvalidAlias
	}
	return nil
}
