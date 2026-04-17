package checker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/SandrZeus/ServicePatrol/internal/alertmanager"
	"github.com/SandrZeus/ServicePatrol/internal/db"
)

type Scheduler struct {
	db      *sql.DB
	mu      sync.Mutex
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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopLocked(target.ID)

	ctx, cancel := context.WithCancel(context.Background())
	s.cancels[target.ID] = cancel
	go s.run(ctx, target)
}

func (s *Scheduler) Stop(targetID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopLocked(targetID)
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

			if s.alerter == nil {
				continue
			}

			s.mu.Lock()
			wasFailing := s.failing[target.ID]
			s.mu.Unlock()

			switch {
			case !result.Success && !wasFailing:
				if err := s.alerter.Fire(target); err != nil {
					log.Printf("could not fire alert: %v", err)
				}
				s.mu.Lock()
				s.failing[target.ID] = true
				s.mu.Unlock()
			case result.Success && wasFailing:
				if err := s.alerter.Resolve(target); err != nil {
					log.Printf("could not resolve alert: %v", err)
				}
				s.mu.Lock()
				s.failing[target.ID] = false
				s.mu.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) stopLocked(targetID int) {
	if cancel, exists := s.cancels[targetID]; exists {
		cancel()
		delete(s.cancels, targetID)
	}
}
