package checker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/SandrZeus/ServicePatrol/internal/alertmanager"
	"github.com/SandrZeus/ServicePatrol/internal/db"
)

type Scheduler struct {
	db      *sql.DB
	cancels map[int]context.CancelFunc
	alerter *alertmanager.AlertmanagerClient
	failing map[int]bool
}

func NewScheduler(db *sql.DB, alerter *alertmanager.AlertmanagerClient) *Scheduler {
	return &Scheduler{
		db:      db,
		cancels: make(map[int]context.CancelFunc),
		alerter: alerter,
		failing: make(map[int]bool),
	}
}

func (s *Scheduler) StartAll() error {
	targets, err := db.GetAll(s.db)
	if err != nil {
		return fmt.Errorf("could not get targets for scheduler: %w", err)
	}
	for _, target := range targets {
		if target.Active {
			s.Start(target)
		}
	}

	return nil
}

func (s *Scheduler) Start(target *db.Target) {
	if _, exists := s.cancels[target.ID]; exists {
		s.Stop(target.ID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancels[target.ID] = cancel
	go s.run(ctx, target)
}

func (s *Scheduler) Stop(targetID int) {
	cancel, exists := s.cancels[targetID]
	if exists {
		cancel()
		delete(s.cancels, targetID)
	}
}

func (s *Scheduler) run(ctx context.Context, target *db.Target) {
	ticker := time.NewTicker(time.Duration(target.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			result := CheckTarget(target)
			err := db.AddCheck(s.db, result)
			if err != nil {
				log.Printf("could not add a check in scheduler: %v", err)
			}
			alert := s.alerter
			if alert != nil {
				if !result.Success && !s.failing[target.ID] {
					if err := s.alerter.Fire(target); err != nil {
						log.Printf("could not fire alert: %v", err)
					}
					s.failing[target.ID] = true
				} else if result.Success && s.failing[target.ID] {
					if err := s.alerter.Resolve(target); err != nil {
						log.Printf("could not resolve alert: %v", err)
					}
					s.failing[target.ID] = false
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
