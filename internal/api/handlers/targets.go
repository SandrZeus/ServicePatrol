package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/SandrZeus/ServicePatrol/internal/checker"
	"github.com/SandrZeus/ServicePatrol/internal/db"
)

type TargetHandler struct {
	db        *sql.DB
	scheduler *checker.Scheduler
}

func NewTargetHandler(db *sql.DB, scheduler *checker.Scheduler) *TargetHandler {
	return &TargetHandler{
		db:        db,
		scheduler: scheduler,
	}
}

func (h *TargetHandler) GetAllTargets(w http.ResponseWriter, r *http.Request) {
	targets, err := db.GetAll(h.db)
	if err != nil {
		http.Error(w, "could not get targets", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(targets); err != nil {
		log.Printf("could not encode response: %v", err)
	}
}

func (h *TargetHandler) GetTargetByID(w http.ResponseWriter, r *http.Request) {
	extract := strings.TrimPrefix(r.URL.Path, "/api/targets/")

	id, err := strconv.Atoi(extract)
	if err != nil {
		http.Error(w, "could not convert string to int", http.StatusBadRequest)
		return
	}

	target, err := db.GetService(h.db, id)
	if err != nil {
		http.Error(w, "target not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(target); err != nil {
		log.Printf("could not encode response: %v", err)
	}
}

func (h *TargetHandler) CreateTarget(w http.ResponseWriter, r *http.Request) {
	var t db.Target
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	target, err := db.AddService(h.db, &t)
	if err != nil {
		http.Error(w, "could not create a service", http.StatusBadRequest)
		return
	}

	if target.Active {
		h.scheduler.Start(target)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(target)
}

func (h *TargetHandler) UpdateTarget(w http.ResponseWriter, r *http.Request) {
	extract := strings.TrimPrefix(r.URL.Path, "/api/targets/")
	id, err := strconv.Atoi(extract)
	if err != nil {
		http.Error(w, "could not convert string to int", http.StatusBadRequest)
		return
	}

	var t db.Target
	err = json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	t.ID = id

	target, err := db.UpdateService(h.db, &t)
	if err != nil {
		http.Error(w, "could not update a service", http.StatusBadRequest)
		return
	}

	h.scheduler.Stop(target.ID)
	if target.Active {
		h.scheduler.Start(target)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(target)
}

func (h *TargetHandler) DeleteTarget(w http.ResponseWriter, r *http.Request) {
	extract := strings.TrimPrefix(r.URL.Path, "/api/targets/")
	id, err := strconv.Atoi(extract)
	if err != nil {
		http.Error(w, "could not convert string to int", http.StatusBadRequest)
		return
	}

	h.scheduler.Stop(id)
	err = db.DeleteService(h.db, id)
	if err != nil {
		http.Error(w, "could not delete target", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
