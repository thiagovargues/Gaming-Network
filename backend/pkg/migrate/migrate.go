package migrate

import (
	"errors"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Apply(dbPath string) error {
	absMigrations, err := filepath.Abs("migrations/sqlite")
	if err != nil {
		return err
	}
	absDB, err := filepath.Abs(dbPath)
	if err != nil {
		return err
	}

	sourceURL := "file://" + absMigrations
	dbURL := "sqlite3://file:" + absDB + "?_foreign_keys=on"

	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
