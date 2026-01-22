package config

import "os"

type Config struct {
	Port               string
	DBPath             string
	MediaDir           string
	CookieName         string
	CookieSecure       bool
	CORSOrigin         string
	FrontendURL        string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	Env                string
}

func Load() Config {
	return Config{
		Port:               getenv("PORT", "8080"),
		DBPath:             getenv("DB_PATH", "storage/app.db"),
		MediaDir:           getenv("MEDIA_DIR", "storage/media"),
		CookieName:         getenv("COOKIE_NAME", "sid"),
		CookieSecure:       getenv("COOKIE_SECURE", "false") == "true",
		CORSOrigin:         getenv("CORS_ORIGIN", "http://localhost:3000"),
		FrontendURL:        getenv("FRONTEND_URL", ""),
		GoogleClientID:     getenv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getenv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getenv("GOOGLE_REDIRECT_URL", ""),
		Env:                getenv("APP_ENV", "dev"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
