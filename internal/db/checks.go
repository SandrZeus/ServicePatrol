package db

import (
	"database/sql"
	"fmt"
	"time"
)

type Check struct {
	ID             int       `json:"id"`
	TargetID       int       `json:"target_id"`
	StatusCode     int       `json:"status_code"`
	ResponseTimeMS int       `json:"response_time_ms"`
	Success        bool      `json:"success"`
	ErrorMessage   *string   `json:"error_message"`
	CheckedAt      time.Time `json:"checked_at"`
}

func AddCheck(db *sql.DB, c *Check) error {
	query := `INSERT INTO check_results (target_id, status_code, response_time_ms, success, error_message, checked_at) VALUES (?, ?, ?, ?, ?, ?);`
	_, err := db.Exec(query, c.TargetID, c.StatusCode, c.ResponseTimeMS, c.Success, c.ErrorMessage, c.CheckedAt)
	if err != nil {
		return fmt.Errorf("could not add check to db: %w", err)
	}

	return nil
}

func GetChecksByTarget(db *sql.DB, id, limit int) ([]*Check, error) {
	query := `SELECT id, target_id, status_code, response_time_ms, success, error_message, checked_at FROM check_results WHERE target_id = ? ORDER BY checked_at DESC LIMIT ?;`
	rows, err := db.Query(query, id, limit)
	if err != nil {
		return nil, fmt.Errorf("could not query db: %w", err)
	}
	defer rows.Close()

	var checks []*Check
	for rows.Next() {
		var c Check
		err := rows.Scan(
			&c.ID,
			&c.TargetID,
			&c.StatusCode,
			&c.ResponseTimeMS,
			&c.Success,
			&c.ErrorMessage,
			&c.CheckedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan a check row: %w", err)
		}
		checks = append(checks, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating check rows: %w", err)
	}

	return checks, nil
}
