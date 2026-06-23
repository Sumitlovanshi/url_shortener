package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Sumitlovanshi/url_shortener/internal/config"
	"github.com/Sumitlovanshi/url_shortener/internal/httpapi"
	"github.com/Sumitlovanshi/url_shortener/internal/shortener"
	"github.com/Sumitlovanshi/url_shortener/internal/store"
)

func main() {
	cfg := config.Load()

	repo, err := store.OpenBbolt(cfg.DataPath)
	if err != nil {
		slog.Error("open datastore", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			slog.Error("close datastore", "error", err)
		}
	}()

	service := shortener.NewService(repo, shortener.NewRandomGenerator(8))

	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("starting url shortener", "addr", addr, "data_path", cfg.DataPath, "base_url", cfg.BaseURL)
	if err := http.ListenAndServe(addr, httpapi.NewHandler(service, cfg.BaseURL)); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
