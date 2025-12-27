package database

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID         string
	Name       string
	KeyPrefix  string // "bm_live_abc12345"
	Scopes     string // comma-separated: "videos:read,videos:write"
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	CreatedAt  time.Time
}

// generate a new API key and returns it (plain key only shown once)
func (db *DB) CreateAPIKey(name, scopes string, expiresAt *time.Time) (*APIKey, string, error) {
	id := uuid.New().String()

	keyBytes := make([]byte, 24)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, "", err
	}

	// PREFIX IS HERE
	plainKey := "bm_" + hex.EncodeToString(keyBytes) // e.g., "bm_a1b2c3..."
	keyPrefix := plainKey[:16]

	keyHash := sha256.Sum256([]byte(plainKey))
	keyHashStr := hex.EncodeToString(keyHash[:])
	query := `INSERT INTO api_keys (id, name, key_hash, key_prefix, scopes, expires_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, id, name, keyHashStr, keyPrefix, scopes, expiresAt)
	if err != nil {
		return nil, "", err
	}
	return &APIKey{
		ID:        id,
		Name:      name,
		KeyPrefix: keyPrefix,
		Scopes:    scopes,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, plainKey, nil
}

// check if API key is valid and returns key info
func (db *DB) ValidateAPIKey(plainKey string) (*APIKey, error) {
	keyHash := sha256.Sum256([]byte(plainKey))
	keyHashStr := hex.EncodeToString(keyHash[:])
	query := `
		SELECT id, name, key_prefix, scopes, last_used_at, expires_at, created_at
		FROM api_keys
		WHERE key_hash = ? AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
	`
	var k APIKey
	err := db.QueryRow(query, keyHashStr).Scan(
		&k.ID, &k.Name, &k.KeyPrefix, &k.Scopes, &k.LastUsedAt, &k.ExpiresAt, &k.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	db.Exec(`UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE id = ?`, k.ID)

	return &k, nil
}

// return all API keys
func (db *DB) ListAPIKeys() ([]APIKey, error) {
	query := `SELECT id, name, key_prefix, scopes, last_used_at, expires_at, created_at FROM api_keys ORDER BY created_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []APIKey
	for rows.Next() {
		var k APIKey
		rows.Scan(&k.ID, &k.Name, &k.KeyPrefix, &k.Scopes, &k.LastUsedAt, &k.ExpiresAt, &k.CreatedAt)
		keys = append(keys, k)
	}
	return keys, nil
}

// delete an API key
func (db *DB) DeleteAPIKey(id string) error {
	query := `DELETE FROM api_keys WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}
