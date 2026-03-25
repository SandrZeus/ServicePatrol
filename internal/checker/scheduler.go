package checker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/SandrZeus/ServicePatrol/internal/db"
)

type Scheduler struct {
	db      *sql.DB
	cancels map[int]context.CancelFunc
}

func NewScheduler(db *sql.DB) *Scheduler {
	return &Scheduler{
		db:      db,
		cancels: make(map[int]context.CancelFunc),
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
		case <-ctx.Done():
			return
		}
	}
}
