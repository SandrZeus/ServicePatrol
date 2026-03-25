package alertmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SandrZeus/ServicePatrol/internal/db"
)

type AlertmanagerClient struct {
	AlertmanagerURL string
}

func NewAlertmanagerClient(URL string) *AlertmanagerClient {
	return &AlertmanagerClient{
		AlertmanagerURL: URL,
	}
}

type alert struct {
	Labels map[string]string `json:"labels"`
	Status string            `json:"status"`
}

func (a *AlertmanagerClient) sendAlert(target *db.Target, status string) error {
	alertStruct := alert{
		Labels: map[string]string{
			"alertname": "TargetDown",
			"target":    target.Name,
			"severity":  "critical",
		},
		Status: status,
	}
	jsonData, err := json.Marshal([]alert{alertStruct})
	if err != nil {
		return fmt.Errorf("could not convert to json: %w", err)
	}

	resp, err := http.Post(a.AlertmanagerURL+"/api/v1/alerts", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("could not make a post request: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func (a *AlertmanagerClient) Fire(target *db.Target) error {
	return a.sendAlert(target, "firing")
}

func (a *AlertmanagerClient) Resolve(target *db.Target) error {
	return a.sendAlert(target, "resolved")
}
