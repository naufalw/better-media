package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)

	if err != nil {
		log.Printf("Error making sqlite %v", err)
		return nil, err
	}

	return &DB{db}, nil
}
