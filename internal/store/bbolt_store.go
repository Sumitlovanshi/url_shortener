package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/Sumitlovanshi/url_shortener/internal/shortener"
	bolt "go.etcd.io/bbolt"
)

var (
	linksByCodeBucket = []byte("links_by_code")
	codeByURLBucket   = []byte("code_by_url")
)

type BboltStore struct {
	db *bolt.DB
}

func OpenBbolt(path string) (*BboltStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}

	store := &BboltStore{db: db}
	if err := store.ensureBuckets(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *BboltStore) Save(ctx context.Context, link shortener.Link) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		links := tx.Bucket(linksByCodeBucket)
		if links.Get([]byte(link.Code)) != nil {
			return shortener.ErrCodeExists
		}

		if !link.CustomAlias {
			urls := tx.Bucket(codeByURLBucket)
			if urls.Get([]byte(link.CanonicalURL)) != nil {
				return shortener.ErrURLExists
			}
		}

		encoded, err := json.Marshal(link)
		if err != nil {
			return err
		}
		if err := links.Put([]byte(link.Code), encoded); err != nil {
			return err
		}

		if !link.CustomAlias {
			urls := tx.Bucket(codeByURLBucket)
			if err := urls.Put([]byte(link.CanonicalURL), []byte(link.Code)); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *BboltStore) FindCodeByCanonicalURL(ctx context.Context, canonicalURL string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	var code string
	err := s.db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(codeByURLBucket).Get([]byte(canonicalURL))
		if value == nil {
			return shortener.ErrNotFound
		}
		code = string(value)
		return nil
	})
	if err != nil {
		return "", err
	}
	return code, nil
}

func (s *BboltStore) GetByCode(ctx context.Context, code string) (shortener.Link, error) {
	if err := ctx.Err(); err != nil {
		return shortener.Link{}, err
	}

	var link shortener.Link
	err := s.db.View(func(tx *bolt.Tx) error {
		return decodeLink(tx.Bucket(linksByCodeBucket).Get([]byte(code)), &link)
	})
	if err != nil {
		return shortener.Link{}, err
	}
	return link, nil
}

func (s *BboltStore) IncrementClick(ctx context.Context, code string) (shortener.Link, error) {
	if err := ctx.Err(); err != nil {
		return shortener.Link{}, err
	}

	var link shortener.Link
	err := s.db.Update(func(tx *bolt.Tx) error {
		links := tx.Bucket(linksByCodeBucket)
		if err := decodeLink(links.Get([]byte(code)), &link); err != nil {
			return err
		}

		link.ClickCount++
		encoded, err := json.Marshal(link)
		if err != nil {
			return err
		}
		return links.Put([]byte(code), encoded)
	})
	if err != nil {
		return shortener.Link{}, err
	}
	return link, nil
}

func (s *BboltStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *BboltStore) ensureBuckets() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(linksByCodeBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(codeByURLBucket); err != nil {
			return err
		}
		return nil
	})
}

func decodeLink(value []byte, link *shortener.Link) error {
	if value == nil {
		return shortener.ErrNotFound
	}
	return json.Unmarshal(value, link)
}
