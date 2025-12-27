package database

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Session struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// create a new user
func (db *DB) CreateUser(email, password, name string) (*User, error) {
	id := uuid.New().String()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	query := `INSERT INTO users (id, email, password_hash, name) VALUES (?, ?, ?, ?)`
	_, err = db.Exec(query, id, email, string(hash), name)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:        id,
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
	}, nil
}

// get user by email address
func (db *DB) GetUserByEmail(email string) (*User, error) {
	query := `SELECT id, email, password_hash, name, role, created_at, updated_at FROM users WHERE email = ?`
	var u User
	err := db.QueryRow(query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// get user by user id
func (db *DB) GetUserByID(id string) (*User, error) {
	query := `SELECT id, email, password_hash, name, role, created_at, updated_at FROM users WHERE id = ?`
	var u User
	err := db.QueryRow(query, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// check if provided password the same as the stored hash
func VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// create new session for a user id
func (db *DB) CreateSession(userID string, duration time.Duration) (*Session, error) {
	id := uuid.New().String()

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes)

	tokenHash := sha256.Sum256([]byte(token))
	tokenHashStr := hex.EncodeToString(tokenHash[:])
	expiresAt := time.Now().Add(duration)
	query := `INSERT INTO sessions (id, user_id, token_hash, expires_at) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, id, userID, tokenHashStr, expiresAt)
	if err != nil {
		return nil, err
	}
	return &Session{
		ID:        id,
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

// check session token if valid and return User
func (db *DB) ValidateSession(token string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(token))
	tokenHashStr := hex.EncodeToString(tokenHash[:])
	query := `
		SELECT u.id, u.email, u.password_hash, u.name, u.created_at, u.updated_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token_hash = ? AND s.expires_at > CURRENT_TIMESTAMP
	`
	var u User
	err := db.QueryRow(query, tokenHashStr).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// Remove a session by token
func (db *DB) DeleteSession(token string) error {
	tokenHash := sha256.Sum256([]byte(token))
	tokenHashStr := hex.EncodeToString(tokenHash[:])
	query := `DELETE FROM sessions WHERE token_hash = ?`
	_, err := db.Exec(query, tokenHashStr)
	return err
}

// Returns number of users in DB
func (db *DB) CountUsers() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// Create a user with admin role
func (db *DB) CreateAdmin(email, password, name string) (*User, error) {
	id := uuid.New().String()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	query := `INSERT INTO users (id, email, password_hash, name, role) VALUES (?, ?, ?, ?, 'admin')`
	_, err = db.Exec(query, id, email, string(hash), name)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:        id,
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
	}, nil
}

// Get all users
func (db *DB) ListUsers() ([]User, error) {
	query := `SELECT id, email, password_hash, name, role, created_at, updated_at FROM users ORDER BY created_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
		users = append(users, u)
	}
	return users, nil
}

// change role
func (db *DB) UpdateUserRole(id, role string) error {
	query := `UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, role, id)
	return err
}

// remove usr
func (db *DB) DeleteUser(id string) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

// change user pw
func (db *DB) UpdateUserPassword(id, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	query := `UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = db.Exec(query, string(hash), id)
	return err
}
