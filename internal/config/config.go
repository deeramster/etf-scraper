package config

import (
	"os"
)

// Config содержит конфигурацию приложения
type Config struct {
	DBPath     string
	ServerPort string
	ScraperURL string
	Verbose    bool
	StaticDir  string
}

// NewConfig создает новую конфигурацию с значениями по умолчанию
func NewConfig() *Config {
	return &Config{
		DBPath:     getEnv("DB_PATH", "etf_data.db"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ScraperURL: getEnv("SCRAPER_URL", "https://assetallocation.ru/etf/"),
		Verbose:    getEnv("VERBOSE", "false") == "true",
		StaticDir:  getEnv("STATIC_DIR", "./static"),
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
