package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

// New opens a SQLite database at dbPath, applies required schema migrations, and returns a DB wrapper.
// It returns an error if opening the database or applying migrations fails.
func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)

	if err != nil {
		log.Printf("Error making sqlite %v", err)
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// migrate ensures the database contains a `jobs` table with the required columns
// and default values. It returns any error encountered while applying the schema.
func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		video_id TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		error TEXT,
		progress INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.Exec(schema)
	return err
}
