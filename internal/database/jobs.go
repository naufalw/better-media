package database

import "time"

type Job struct {
	ID        string
	VideoID   string
	Status    string // pending, processing, playable, completed, failed
	Error     *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// insert a new job with status "pending"
func (db *DB) CreateJob(id, videoID string) error {
	mutation := `
		INSERT INTO jobs (id, video_id, status) VALUES (?, ?, ?)
	`

	_, err := db.Exec(mutation, id, videoID, "pending")
	return err
}

// fetch a job by ID
func (db *DB) GetJob(id string) (*Job, error) {
	query := `SELECT id, video_id, status, error, created_at, updated_at 
              FROM jobs WHERE id = ?`

	row := db.QueryRow(query, id)

	var job Job

	err := row.Scan(
		&job.ID,
		&job.VideoID,
		&job.Status,
		&job.Error,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &job, nil
}

// change the job status
func (db *DB) UpdateJobStatus(id, status string, errMsg *string) error {
	query := `UPDATE jobs 
              SET status = ?, error = ?, updated_at = CURRENT_TIMESTAMP 
              WHERE id = ?`

	_, err := db.Exec(query, status, errMsg, id)
	return err
}
