package config

import "os"

type Config struct {
	Port     string
	BaseURL  string
	DataPath string
}

func Load() Config {
	return Config{
		Port:     getEnv("PORT", "8080"),
		BaseURL:  getEnv("BASE_URL", "http://localhost:8080"),
		DataPath: getEnv("DATA_PATH", "data/url_shortener.db"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
