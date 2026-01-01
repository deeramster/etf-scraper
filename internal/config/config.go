package config

import (
	"os"
	"strings"
)

type Config struct {
	DBPath          string
	ServerPort      string
	AdminPort       string
	ScraperURL      string
	Verbose         bool
	StaticDir       string
	CACertPath      string
	ServerCertPath  string
	ServerKeyPath   string
	AdminAllowedDNs []string
}

func NewConfig() *Config {
	allowedDNs := getEnv("ADMIN_ALLOWED_DNS", "CN=adminetf,OU=Administrators,O=admins,L=Moscow,ST=Moscow,C=RU")
	var dnList []string
	if allowedDNs != "" {
		dnList = strings.Split(allowedDNs, ";")
		for i := range dnList {
			dnList[i] = strings.TrimSpace(dnList[i])
		}
	}

	return &Config{
		DBPath:          getEnv("DB_PATH", "etf_data.db"),
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		AdminPort:       getEnv("ADMIN_PORT", "8443"),
		ScraperURL:      getEnv("SCRAPER_URL", "https://assetallocation.ru/etf/"),
		Verbose:         getEnv("VERBOSE", "false") == "true",
		StaticDir:       getEnv("STATIC_DIR", "./static"),
		CACertPath:      getEnv("CA_CERT_PATH", "./certs/ca.crt"),
		ServerCertPath:  getEnv("SERVER_CERT_PATH", "./certs/server.crt"),
		ServerKeyPath:   getEnv("SERVER_KEY_PATH", "./certs/server.key"),
		AdminAllowedDNs: dnList,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
