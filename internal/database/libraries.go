package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Library struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (db *DB) CreateLibrary(name, description string) (*Library, error) {
	id := uuid.New().String()
	query := `INSERT INTO libraries (id, name, description) VALUES (?, ?, ?)`
	_, err := db.Exec(query, id, name, description)
	if err != nil {
		return nil, err
	}
	return &Library{ID: id, Name: name, Description: description, CreatedAt: time.Now()}, nil
}

func (db *DB) GetLibrary(id string) (*Library, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM libraries WHERE id = ?`
	var l Library
	err := db.QueryRow(query, id).Scan(&l.ID, &l.Name, &l.Description, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (db *DB) ListLibraries() ([]Library, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM libraries ORDER BY created_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var libs []Library
	for rows.Next() {
		var l Library
		rows.Scan(&l.ID, &l.Name, &l.Description, &l.CreatedAt, &l.UpdatedAt)
		libs = append(libs, l)
	}
	return libs, nil
}

func (db *DB) UpdateLibrary(id, name, description string) error {
	query := `UPDATE libraries SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, name, description, id)
	return err
}

func (db *DB) DeleteLibrary(id string) error {
	query := `DELETE FROM libraries WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}
