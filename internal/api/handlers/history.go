package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/SandrZeus/ServicePatrol/internal/db"
)

type HistoryHandler struct {
	db *sql.DB
}

func NewHistoryHandler(db *sql.DB) *HistoryHandler {
	return &HistoryHandler{db: db}
}

func (h *HistoryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	extract := strings.TrimPrefix(r.URL.Path, "/api/targets/")
	parts := strings.Split(extract, "/")
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "could not get a target id", http.StatusBadRequest)
		return
	}

	limit := 50
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
	}

	history, err := db.GetChecksByTarget(h.db, id, limit)
	if err != nil {
		http.Error(w, "could not get target history", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(history)
}
