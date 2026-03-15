package storage

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // Pure-Go SQLite driver
)

// NewSQLiteDB opens (or creates) a SQLite database at the given path and runs
// auto-migrations to ensure the schema is up to date.
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

// migrate runs all schema migrations.
func migrate(db *sql.DB) error {
	// Drop old table to migrate clean from Phase 1
	db.Exec("DROP TABLE IF EXISTS etfs")

	const createContracts = `
CREATE TABLE IF NOT EXISTS contracts (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    conid        INTEGER NOT NULL UNIQUE,
    symbol       TEXT    NOT NULL,
    company_name TEXT    NOT NULL,
    exchange     TEXT    NOT NULL
);`

	if _, err := db.Exec(createContracts); err != nil {
		return fmt.Errorf("create contracts table: %w", err)
	}
	return nil
}
