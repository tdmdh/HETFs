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
	const createETFs = `
CREATE TABLE IF NOT EXISTS etfs (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol   TEXT    NOT NULL,
    isin     TEXT    NOT NULL,
    exchange TEXT    NOT NULL,
    currency TEXT    NOT NULL,
    UNIQUE(isin, exchange)
);`

	if _, err := db.Exec(createETFs); err != nil {
		return fmt.Errorf("create etfs table: %w", err)
	}
	return nil
}
