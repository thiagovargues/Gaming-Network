package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"backend/internal/config"
	httpapi "backend/internal/http"
	"backend/internal/repo/sqlite"
	"backend/pkg/migrate"
)

func main() {
	cfg := config.Load()

	database, err := sqlite.OpenSQLite(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	if err := migrate.Apply(cfg.DBPath); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      httpapi.NewRouter(database),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server: %v", err)
	}
}
