package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"backend/internal/db"
	"backend/internal/httpapi"
)

func main() {
	port := envOrDefault("PORT", "8080")
	dbPath := envOrDefault("DB_PATH", "data/app.db")

	database, err := db.OpenSQLite(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	if err := db.ApplyMigrations(dbPath); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      httpapi.NewRouter(database),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
