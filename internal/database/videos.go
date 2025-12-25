package database

import "time"

type Video struct {
	ID        string
	Title     string
	Status    string
	Source    string
	CreatedAt time.Time
}

// Add new video record
func (db *DB) CreateVideo(id, title, source string) error {
	query := `INSERT INTO videos (id, title, status, source) VALUES (?, ?, 'processing', ?)`
	_, err := db.Exec(query, id, title, source)
	return err
}

// Update  video status
func (db *DB) UpdateVideoStatus(id, status string) error {
	query := `UPDATE videos SET status = ? WHERE id = ?`
	_, err := db.Exec(query, status, id)
	return err
}

// Return all videos
func (db *DB) ListVideos() ([]Video, error) {
	query := `SELECT id, title, status, source, created_at FROM videos ORDER BY created_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []Video
	for rows.Next() {
		var v Video
		rows.Scan(&v.ID, &v.Title, &v.Status, &v.Source, &v.CreatedAt)
		videos = append(videos, v)
	}
	return videos, nil
}

// Returns a single video
func (db *DB) GetVideo(id string) (*Video, error) {
	query := `SELECT id, title, status, source, created_at FROM videos WHERE id = ?`
	var v Video
	err := db.QueryRow(query, id).Scan(&v.ID, &v.Title, &v.Status, &v.Source, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Remove a video from DB
func (db *DB) DeleteVideo(id string) error {
	query := `DELETE FROM videos WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}
