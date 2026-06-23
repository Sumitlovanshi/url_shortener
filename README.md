# URL Shortener

A small Go URL shortener with persistent storage, custom aliases, redirects, and minimal link analytics.

## Features

- `POST /shorten` creates a short code for an absolute `http` or `https` URL.
- `GET /{code}` redirects to the original URL with `301 Moved Permanently`.
- `GET /links/{code}/stats` returns the redirect count and link metadata.
- Custom aliases are supported with safe URL characters.
- Repeated generated shorten requests for the same canonical URL return the existing code.
- Mappings are persisted in an embedded bbolt database.

## Requirements

- Go 1.23+
- Docker 29+ if running with Docker

## Run Locally

```bash
make run
```

The server starts on `http://localhost:8080` by default.

Configuration is environment based:

```bash
PORT=8080
BASE_URL=http://localhost:8080
DATA_PATH=data/url_shortener.db
```

## Run With Docker

```bash
docker compose up --build
```

The compose file stores the bbolt database in a named Docker volume.

## Test

```bash
make test
make race
make vet
```

## API Examples

Create a generated short link:

```bash
curl -s -X POST http://localhost:8080/shorten \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/some/long/path"}'
```

Create a custom alias:

```bash
curl -s -X POST http://localhost:8080/shorten \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/docs","alias":"docs"}'
```

Redirect:

```bash
curl -i http://localhost:8080/docs
```

Stats:

```bash
curl -s http://localhost:8080/links/docs/stats
```

## Design Notes

The service uses `crypto/rand` to generate Base62 codes with a default length of 8. Each generated candidate is checked in a bbolt write transaction before it is accepted, so collisions are handled by retrying rather than assuming probability is enough.

Duplicate generated requests are idempotent: after URL canonicalization, the same URL returns the existing code with `"reused": true`. Custom aliases are treated as user-chosen identifiers and do not replace the generated mapping for that URL.

The storage implementation sits behind the `shortener.Repository` interface. That keeps the current embedded bbolt choice small while leaving room to swap in PostgreSQL or Redis later without changing HTTP handlers or core business logic.

## Known Limitations

- No authentication or rate limiting.
- No link expiration.
- Analytics are intentionally minimal: only a click counter is stored.
- bbolt is single-node embedded storage, not a horizontally scalable datastore.
