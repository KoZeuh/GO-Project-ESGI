// Package database gère la connexion SQLite et l'application du schéma.
package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // driver pur Go, aucune dépendance CGO requise
)

var schema string

// Connect ouvre la base SQLite située à dsn et applique le schéma (idempotent)
func Connect(dsn string) (*sql.DB, error) {
	if dir := filepath.Dir(dsn); dsn != ":memory:" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("création du dossier de la base %q : %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("ouverture de la base %q : %w", dsn, err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("activation des foreign keys : %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("application du schéma : %w", err)
	}

	return db, nil
}
