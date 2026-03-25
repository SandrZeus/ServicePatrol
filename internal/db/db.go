package db

import (
	"database/sql"
	"fmt"
)

func Init(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.Exec("PRAGMA foreign_keys = ON")

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS targets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		method TEXT NOT NULL DEFAULT 'GET',
		interval_seconds INTEGER NOT NULL DEFAULT 30,
		timeout_seconds INTEGER NOT NULL DEFAULT 5,
		expected_status INTEGER NOT NULL DEFAULT 200,
		active BOOLEAN NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS check_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_id INTEGER NOT NULL,
		status_code INTEGER,
		response_time_ms INTEGER,
		success BOOLEAN NOT NULL,
		error_message TEXT,
		checked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_check_results_target_id ON check_results(target_id);
	CREATE INDEX IF NOT EXISTS idx_check_results_checked_at ON check_results(checked_at);
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	return nil
}
