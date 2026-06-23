package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/Sumitlovanshi/url_shortener/internal/shortener"
)

type Handler struct {
	service *shortener.Service
	baseURL string
}

func NewHandler(service *shortener.Service, baseURL string) http.Handler {
	handler := &Handler{
		service: service,
		baseURL: strings.TrimRight(baseURL, "/"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.health)
	mux.HandleFunc("POST /shorten", handler.shorten)
	mux.HandleFunc("GET /links/{code}/stats", handler.stats)
	mux.HandleFunc("GET /{code}", handler.redirect)

	return withRecovery(mux)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	var request shortenRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json request body")
		return
	}

	result, err := h.service.Shorten(r.Context(), shortener.ShortenInput{
		URL:   request.URL,
		Alias: request.Alias,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	status := http.StatusCreated
	if result.Reused {
		status = http.StatusOK
	}

	writeJSON(w, status, shortenResponse{
		Code:     result.Link.Code,
		ShortURL: h.shortURL(result.Link.Code),
		URL:      result.Link.OriginalURL,
		Reused:   result.Reused,
	})
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request) {
	link, err := h.service.RecordRedirect(r.Context(), r.PathValue("code"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	http.Redirect(w, r, link.OriginalURL, http.StatusMovedPermanently)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	link, err := h.service.GetStats(r.Context(), r.PathValue("code"))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statsResponse{
		Code:        link.Code,
		URL:         link.OriginalURL,
		ClickCount:  link.ClickCount,
		CreatedAt:   link.CreatedAt,
		CustomAlias: link.CustomAlias,
	})
}

func (h *Handler) shortURL(code string) string {
	return fmt.Sprintf("%s/%s", h.baseURL, url.PathEscape(code))
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shortener.ErrInvalidURL):
		writeError(w, http.StatusBadRequest, "invalid url; use an absolute http or https URL")
	case errors.Is(err, shortener.ErrInvalidAlias):
		writeError(w, http.StatusBadRequest, "invalid alias; use 3-64 letters, digits, underscores, or hyphens")
	case errors.Is(err, shortener.ErrAliasConflict):
		writeError(w, http.StatusConflict, "alias already exists")
	case errors.Is(err, shortener.ErrNotFound):
		writeError(w, http.StatusNotFound, "short code not found")
	default:
		slog.Error("request failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("write json response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
