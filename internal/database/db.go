package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

// Open a SQLite database at dbPath, applies required schema migrations, and returns a DB wrapper.
func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Printf("Error opening sqlite: %v", err)
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// migrate applies all schema migrations
func migrate(db *sql.DB) error {
	schema := `
	-- Users
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Organizations 
	CREATE TABLE IF NOT EXISTS organizations (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Organization members 
	CREATE TABLE IF NOT EXISTS organization_members (
		organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role TEXT NOT NULL DEFAULT 'member', -- 'owner', 'admin', 'member'
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (organization_id, user_id)
	);

	-- Library
	CREATE TABLE IF NOT EXISTS libraries (
		id TEXT PRIMARY KEY,
		organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Videos
	CREATE TABLE IF NOT EXISTS videos (
		id TEXT PRIMARY KEY,
		library_id TEXT REFERENCES libraries(id) ON DELETE SET NULL,
		title TEXT,
		description TEXT,
		status TEXT DEFAULT 'processing',
		source TEXT, -- 'upload', 'livestream'
		duration_ms INTEGER,
		file_size_bytes INTEGER,
		resolution_width INTEGER,
		resolution_height INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Stream key 
	CREATE TABLE IF NOT EXISTS stream_keys (
		id TEXT PRIMARY KEY,
		library_id TEXT REFERENCES libraries(id) ON DELETE SET NULL,
		name TEXT NOT NULL,
		key TEXT UNIQUE NOT NULL,
		active INTEGER DEFAULT 1,
		save_recording INTEGER DEFAULT 1, -- whether to save livestream recordings
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Transcoding job
	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		video_id TEXT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
		status TEXT NOT NULL DEFAULT 'pending',
		error TEXT,
		progress INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- API Keys 
	CREATE TABLE IF NOT EXISTS api_keys (
		id TEXT PRIMARY KEY,
		organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		key_hash TEXT UNIQUE NOT NULL, -- hashed key, actual key shown only on creation
		key_prefix TEXT NOT NULL, -- first 8 chars for identification (e.g., "bm_live_")
		scopes TEXT, -- comma-separated: 'videos:read', 'videos:write', 'streams:read', etc.
		last_used_at DATETIME,
		expires_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- User session
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- VOD Transcoding Settings
	CREATE TABLE IF NOT EXISTS transcoding_presets (
		id TEXT PRIMARY KEY,
		organization_id TEXT REFERENCES organizations(id) ON DELETE CASCADE, -- NULL = global/system preset
		name TEXT NOT NULL,
		resolutions TEXT NOT NULL, -- JSON: ["1080p", "720p", "480p"]
		video_codec TEXT DEFAULT 'h264',
		audio_codec TEXT DEFAULT 'aac',
		is_default INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_videos_library ON videos(library_id);
	CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);
	CREATE INDEX IF NOT EXISTS idx_stream_keys_library ON stream_keys(library_id);
	CREATE INDEX IF NOT EXISTS idx_jobs_video ON jobs(video_id);
	CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
	CREATE INDEX IF NOT EXISTS idx_api_keys_org ON api_keys(organization_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_libraries_org ON libraries(organization_id);
	CREATE INDEX IF NOT EXISTS idx_org_members_user ON organization_members(user_id);
	`

	_, err := db.Exec(schema)
	return err
}
