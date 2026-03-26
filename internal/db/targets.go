package db

import (
	"database/sql"
	"fmt"
	"time"
)

type Target struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	URL             string    `json:"url"`
	Method          string    `json:"method"`
	IntervalSeconds int       `json:"interval_seconds"`
	TimeoutSeconds  int       `json:"timeout_seconds"`
	ExpectedStatus  int       `json:"expected_status"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func GetAll(db *sql.DB) ([]*Target, error) {
	query := `SELECT id, name, url, method, interval_seconds, timeout_seconds, expected_status, active, created_at, updated_at FROM targets;`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("could not get rows: %w", err)
	}
	defer rows.Close()

	var targets []*Target
	for rows.Next() {
		var t Target
		err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.URL,
			&t.Method,
			&t.IntervalSeconds,
			&t.TimeoutSeconds,
			&t.ExpectedStatus,
			&t.Active,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan a row: %w", err)
		}
		targets = append(targets, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return targets, nil
}

func GetService(db *sql.DB, id int) (*Target, error) {
	query := `SELECT id, name, url, method, interval_seconds, timeout_seconds, expected_status, active, created_at, updated_at FROM targets WHERE id = ?;`
	row := db.QueryRow(query, id)

	var t Target
	err := row.Scan(
		&t.ID,
		&t.Name,
		&t.URL,
		&t.Method,
		&t.IntervalSeconds,
		&t.TimeoutSeconds,
		&t.ExpectedStatus,
		&t.Active,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("could not scan a row: %w", err)
	}

	return &t, nil
}

func AddService(db *sql.DB, t *Target) (*Target, error) {
	query := `INSERT INTO targets (name, url, method, interval_seconds, timeout_seconds, expected_status, active) VALUES (?, ?, ?, ?, ?, ?, ?);`
	result, err := db.Exec(query, t.Name, t.URL, t.Method, t.IntervalSeconds, t.TimeoutSeconds, t.ExpectedStatus, t.Active)
	if err != nil {
		return nil, fmt.Errorf("could not insert a service: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("could not get a new id: %w", err)
	}

	return GetService(db, int(id))
}

func UpdateService(db *sql.DB, t *Target) (*Target, error) {
	query := `UPDATE targets SET name = ?, url = ?, method = ?, interval_seconds = ?, timeout_seconds = ?, expected_status = ?, active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;`
	_, err := db.Exec(query, t.Name, t.URL, t.Method, t.IntervalSeconds, t.TimeoutSeconds, t.ExpectedStatus, t.Active, t.ID)
	if err != nil {
		return nil, fmt.Errorf("could not update a service: %w", err)
	}

	return GetService(db, t.ID)
}

func DeleteService(db *sql.DB, id int) error {
	query := `DELETE FROM targets WHERE id = ?;`
	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("could not delete a service: %w", err)
	}

	return nil
}
