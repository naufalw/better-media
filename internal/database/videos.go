package database

import "time"

type Video struct {
	ID               string
	LibraryID        *string
	Title            string
	Description      string
	Status           string
	Source           string // "upload", "livestream"
	DurationMs       *int64
	FileSizeBytes    *int64
	ResolutionWidth  *int
	ResolutionHeight *int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Add new video record
func (db *DB) CreateVideo(id, title, source string, libraryID *string) error {
	query := `INSERT INTO videos (id, title, status, source, library_id) VALUES (?, ?, 'processing', ?, ?)`
	_, err := db.Exec(query, id, title, source, libraryID)
	return err
}

// Update  video status
func (db *DB) UpdateVideoStatus(id, status string) error {
	query := `UPDATE videos SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, status, id)
	return err
}

// Update video metadata
func (db *DB) UpdateVideoMetadata(id string, durationMs, fileSizeBytes int64, width, height int) error {
	query := `UPDATE videos SET duration_ms = ?, file_size_bytes = ?, resolution_width = ?, resolution_height = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, durationMs, fileSizeBytes, width, height, id)
	return err
}

// Return all videos
func (db *DB) ListVideos() ([]Video, error) {
	query := `SELECT id, library_id, title, description, status, source, duration_ms, file_size_bytes, resolution_width, resolution_height, created_at, updated_at FROM videos ORDER BY created_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var videos []Video
	for rows.Next() {
		var v Video
		rows.Scan(&v.ID, &v.LibraryID, &v.Title, &v.Description, &v.Status, &v.Source, &v.DurationMs, &v.FileSizeBytes, &v.ResolutionWidth, &v.ResolutionHeight, &v.CreatedAt, &v.UpdatedAt)
		videos = append(videos, v)
	}
	return videos, nil
}

// Return videos by library
func (db *DB) ListVideosByLibrary(libraryID string) ([]Video, error) {
	query := `SELECT id, library_id, title, description, status, source, duration_ms, file_size_bytes, resolution_width, resolution_height, created_at, updated_at FROM videos WHERE library_id = ? ORDER BY created_at DESC`
	rows, err := db.Query(query, libraryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var videos []Video
	for rows.Next() {
		var v Video
		rows.Scan(&v.ID, &v.LibraryID, &v.Title, &v.Description, &v.Status, &v.Source, &v.DurationMs, &v.FileSizeBytes, &v.ResolutionWidth, &v.ResolutionHeight, &v.CreatedAt, &v.UpdatedAt)
		videos = append(videos, v)
	}
	return videos, nil
}

// Returns a single video
func (db *DB) GetVideo(id string) (*Video, error) {
	query := `SELECT id, library_id, title, description, status, source, duration_ms, file_size_bytes, resolution_width, resolution_height, created_at, updated_at FROM videos WHERE id = ?`
	var v Video
	err := db.QueryRow(query, id).Scan(&v.ID, &v.LibraryID, &v.Title, &v.Description, &v.Status, &v.Source, &v.DurationMs, &v.FileSizeBytes, &v.ResolutionWidth, &v.ResolutionHeight, &v.CreatedAt, &v.UpdatedAt)
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

// change which library the video is stored
func (db *DB) UpdateVideoLibrary(videoID string, libraryID *string) error {
	query := `UPDATE videos SET library_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, libraryID, videoID)
	return err
}
