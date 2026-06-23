package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Sumitlovanshi/url_shortener/internal/shortener"
	"github.com/Sumitlovanshi/url_shortener/internal/store"
)

func TestShortenCreatesAndReusesGeneratedCode(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)

	first := postJSON(t, handler, "/shorten", `{"url":"https://example.com/docs"}`)
	if first.Code == "" {
		t.Fatal("code is empty")
	}
	if first.ShortURL != "http://sho.rt/"+first.Code {
		t.Fatalf("short_url = %q, want http://sho.rt/%s", first.ShortURL, first.Code)
	}
	if first.Reused {
		t.Fatal("first Reused = true, want false")
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(`{"url":"https://example.com/docs"}`))
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("duplicate status = %d, want %d, body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var second shortenResponse
	decodeJSON(t, recorder, &second)
	if second.Code != first.Code {
		t.Fatalf("second code = %q, want %q", second.Code, first.Code)
	}
	if !second.Reused {
		t.Fatal("second Reused = false, want true")
	}
}

func TestShortenRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(`{"url":"ftp://example.com/file"}`))

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestShortenRejectsDuplicateAlias(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)
	_ = postJSON(t, handler, "/shorten", `{"url":"https://example.com/one","alias":"my-link"}`)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(`{"url":"https://example.com/two","alias":"my-link"}`))

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
}

func TestRedirectIncrementsStats(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)
	created := postJSON(t, handler, "/shorten", `{"url":"https://example.com/target","alias":"go-test"}`)

	redirectRecorder := httptest.NewRecorder()
	redirectRequest := httptest.NewRequest(http.MethodGet, "/"+created.Code, nil)
	handler.ServeHTTP(redirectRecorder, redirectRequest)

	if redirectRecorder.Code != http.StatusMovedPermanently {
		t.Fatalf("redirect status = %d, want %d", redirectRecorder.Code, http.StatusMovedPermanently)
	}
	if location := redirectRecorder.Header().Get("Location"); location != "https://example.com/target" {
		t.Fatalf("Location = %q, want https://example.com/target", location)
	}

	statsRecorder := httptest.NewRecorder()
	statsRequest := httptest.NewRequest(http.MethodGet, "/links/"+created.Code+"/stats", nil)
	handler.ServeHTTP(statsRecorder, statsRequest)

	if statsRecorder.Code != http.StatusOK {
		t.Fatalf("stats status = %d, want %d", statsRecorder.Code, http.StatusOK)
	}

	var stats statsResponse
	decodeJSON(t, statsRecorder, &stats)
	if stats.ClickCount != 1 {
		t.Fatalf("ClickCount = %d, want 1", stats.ClickCount)
	}
	if !stats.CustomAlias {
		t.Fatal("CustomAlias = false, want true")
	}
}

func TestUnknownCodeReturnsNotFound(t *testing.T) {
	t.Parallel()

	handler := newTestHandler(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/missing", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func postJSON(t *testing.T, handler http.Handler, path string, body string) shortenResponse {
	t.Helper()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	var response shortenResponse
	decodeJSON(t, recorder, &response)
	return response
}

func decodeJSON(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, recorder.Body.String())
	}
}

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()

	repo, err := store.OpenBbolt(filepath.Join(t.TempDir(), "links.db"))
	if err != nil {
		t.Fatalf("OpenBbolt() error = %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	service := shortener.NewService(repo, shortener.NewRandomGenerator(8))
	return NewHandler(service, "http://sho.rt")
}
