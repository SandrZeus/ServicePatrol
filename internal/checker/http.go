package checker

import (
	"net/http"
	"time"

	"github.com/SandrZeus/ServicePatrol/internal/db"
)

func CheckTarget(t *db.Target) *db.Check {
	client := &http.Client{
		Timeout: time.Duration(t.TimeoutSeconds) * time.Second,
	}

	req, err := http.NewRequest(t.Method, t.URL, nil)
	if err != nil {
		errMsg := err.Error()
		return &db.Check{
			TargetID:     t.ID,
			Success:      false,
			ErrorMessage: &errMsg,
			CheckedAt:    time.Now(),
		}
	}

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		errMsg := err.Error()
		return &db.Check{
			TargetID:     t.ID,
			Success:      false,
			ErrorMessage: &errMsg,
			CheckedAt:    time.Now(),
		}
	}
	defer resp.Body.Close()

	duration := time.Since(start).Milliseconds()

	return &db.Check{
		TargetID:       t.ID,
		StatusCode:     resp.StatusCode,
		ResponseTimeMS: int(duration),
		Success:        resp.StatusCode == t.ExpectedStatus,
		ErrorMessage:   nil,
		CheckedAt:      time.Now(),
	}
}
