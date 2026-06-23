package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Sumitlovanshi/url_shortener/internal/config"
	"github.com/Sumitlovanshi/url_shortener/internal/httpapi"
)

func main() {
	cfg := config.Load()

	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("starting url shortener", "addr", addr)
	if err := http.ListenAndServe(addr, httpapi.NewHandler()); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
