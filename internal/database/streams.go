package database

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type StreamKey struct {
	ID        string
	Name      string
	Key       string
	Active    bool
	CreatedAt time.Time
}

// create a random stream key
func GenerateStreamKey() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// creates a new stream key
func (db *DB) CreateStreamKey(name string) (*StreamKey, error) {
	id := GenerateStreamKey()[:8]
	key := GenerateStreamKey()
	query := `INSERT INTO stream_keys (id, name, key, active) VALUES (?, ?, ?, 1)`
	_, err := db.Exec(query, id, name, key)
	if err != nil {
		return nil, err
	}
	return &StreamKey{ID: id, Name: name, Key: key, Active: true}, nil
}

// Check if a stream key is valid and active
func (db *DB) ValidateStreamKey(key string) bool {
	var active int
	query := `SELECT active FROM stream_keys WHERE key = ?`
	err := db.QueryRow(query, key).Scan(&active)
	return err == nil && active == 1
}

// Return all stream keys
func (db *DB) ListStreamKeys() ([]StreamKey, error) {
	query := `SELECT id, name, key, active, created_at FROM stream_keys`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []StreamKey
	for rows.Next() {
		var k StreamKey
		rows.Scan(&k.ID, &k.Name, &k.Key, &k.Active, &k.CreatedAt)
		keys = append(keys, k)
	}
	return keys, nil
}
