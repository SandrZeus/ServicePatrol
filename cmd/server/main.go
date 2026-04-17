package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/SandrZeus/ServicePatrol/internal/alertmanager"
	"github.com/SandrZeus/ServicePatrol/internal/api/handlers"
	"github.com/SandrZeus/ServicePatrol/internal/api/middleware"
	"github.com/SandrZeus/ServicePatrol/internal/checker"
	"github.com/SandrZeus/ServicePatrol/internal/config"
	"github.com/SandrZeus/ServicePatrol/internal/db"
	"github.com/SandrZeus/ServicePatrol/internal/events"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file found, using environment variables")
	}

	cfg := config.Load()

	database, err := db.Init(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("could not close database: %v", err)
		}
	}()

	var alerter *alertmanager.AlertmanagerClient

	if cfg.AlertmanagerToggle {
		alerter = alertmanager.NewAlertmanagerClient(cfg.AlertmanagerURL)
	}

	bus := events.NewBus()
	sched := checker.NewScheduler(database, alerter, bus)
	if err := sched.StartAll(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	th := handlers.NewTargetHandler(database, sched)
	hh := handlers.NewHistoryHandler(database)
	eh := handlers.NewEventsHandler(bus)
	mux.HandleFunc("/api/targets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			th.GetAllTargets(w, r)
		case http.MethodPost:
			th.CreateTarget(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/targets/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/history") {
			hh.GetHistory(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			th.GetTargetByID(w, r)
		case http.MethodPut:
			th.UpdateTarget(w, r)
		case http.MethodDelete:
			th.DeleteTarget(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/events", eh.Stream)

	addr := ":" + cfg.ServerPort
	log.Printf("server started on %s", addr)
	if err := http.ListenAndServe(addr, middleware.CORS(mux, cfg.CORSOrigin)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
