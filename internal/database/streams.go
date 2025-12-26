package database

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type StreamKey struct {
	ID            string
	LibraryID     *string
	Name          string
	Key           string
	Active        bool
	SaveRecording bool
	CreatedAt     time.Time
}

// create a random stream key
func generateStreamKey() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("failed to generate random bytes: " + err.Error())
	}
	return hex.EncodeToString(bytes)
}

// creates a new stream key
func (db *DB) CreateStreamKey(name string, libraryID *string, saveRecording bool) (*StreamKey, error) {
	id := uuid.New().String()
	key := generateStreamKey()
	query := `INSERT INTO stream_keys (id, name, key, library_id, active, save_recording) VALUES (?, ?, ?, ?, 1, ?)`
	_, err := db.Exec(query, id, name, key, libraryID, saveRecording)
	if err != nil {
		return nil, err
	}
	return &StreamKey{
		ID:            id,
		LibraryID:     libraryID,
		Name:          name,
		Key:           key,
		Active:        true,
		SaveRecording: saveRecording,
		CreatedAt:     time.Now(),
	}, nil
}

// Check if a stream key is valid and active
func (db *DB) ValidateStreamKey(key string) (*StreamKey, error) {
	query := `SELECT id, library_id, name, key, active, save_recording, created_at FROM stream_keys WHERE key = ? AND active = 1`
	var s StreamKey
	err := db.QueryRow(query, key).Scan(&s.ID, &s.LibraryID, &s.Name, &s.Key, &s.Active, &s.SaveRecording, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Return all stream keys
func (db *DB) ListStreamKeys() ([]StreamKey, error) {
	query := `SELECT id, library_id, name, key, active, save_recording, created_at FROM stream_keys ORDER BY created_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []StreamKey
	for rows.Next() {
		var k StreamKey
		rows.Scan(&k.ID, &k.LibraryID, &k.Name, &k.Key, &k.Active, &k.SaveRecording, &k.CreatedAt)
		keys = append(keys, k)
	}
	return keys, nil
}

// Deactivate a stream key
func (db *DB) DeactivateStreamKey(id string) error {
	query := `UPDATE stream_keys SET active = 0 WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

// Delete a stream key
func (db *DB) DeleteStreamKey(id string) error {
	query := `DELETE FROM stream_keys WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}
