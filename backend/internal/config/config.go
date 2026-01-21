package config

import "os"

type Config struct {
	Port   string
	DBPath string
}

func Load() Config {
	return Config{
		Port:   getenv("PORT", "8080"),
		DBPath: getenv("DB_PATH", "storage/app.db"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
