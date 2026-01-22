package main

import (
	"log"
	"net/http"
	"time"

	"backend/internal/config"
	apphttp "backend/internal/http"
	"backend/internal/repo/sqlite"
	"backend/pkg/migrate"
)

func main() {
	cfg := config.Load()

	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := migrate.Apply(cfg.DBPath); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	handler := apphttp.NewRouter(cfg, db)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
