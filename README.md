# ServicePatrol

Lightweight HTTP health monitoring service written in Go. Designed for self-hosted environments and homelabs.

ServicePatrol monitors your services by probing HTTP endpoints at configurable intervals, stores check history in SQLite, and optionally fires alerts through Alertmanager when a target goes down or recovers.

## Features

- **Per-target scheduling** — each monitored service runs in its own goroutine with independent intervals and timeouts
- **SQLite persistence** — zero-dependency storage with automatic schema migrations
- **Alertmanager integration** — optional firing and auto-resolving alerts on state changes
- **REST API** — full CRUD for targets and paginated check history, compatible with any frontend
- **Minimal footprint** — single binary, ~5MB memory at runtime
- **Real-time event stream** — Server-Sent Events endpoint broadcasts every check result and state transition as it happens, for live dashboards without polling
- **Concurrent-safe scheduler** — mutex-protected goroutine management supports concurrent CRUD and live reconfiguration without restarts

## Architecture

```
┌──────────────┐      ┌─────────────┐      ┌──────────────────┐
│   REST API   │◄────►│  SQLite DB  │◄────►│    Scheduler     │
│  (handlers)  │      │  (targets,  │      │  (per-target     │
│              │      │   checks)   │      │   goroutines)    │
└──────────────┘      └─────────────┘      └───────┬──────────┘
                                                   │
                                          ┌────────▼─────────┐
                                          │   HTTP Checker   │
                                          │  (probe targets) │
                                          └────────┬─────────┘
                                                   │
                              ┌────────────────────┼──────────────┐
                              │                    │              │
                      ┌───────▼─────────┐  ┌───────▼──────┐  ┌────▼──────────┐
                      │   SQLite DB     │  │  Event Bus   │  │ Alertmanager  │
                      │  (check_results)│  │  (pub/sub)   │  │  (optional)   │
                      └─────────────────┘  └───────┬──────┘  └───────────────┘
                                                   │
                                           ┌───────▼─────────┐
                                           │  SSE Subscribers│
                                           │  (/api/events)  │
                                           └─────────────────┘
```

```
internal/
├── api/
│   ├── handlers/       # REST endpoints for targets, history, and event stream
│   └── middleware/      # CORS
├── alertmanager/        # Alertmanager client (fire/resolve)
├── checker/
│   ├── http.go          # HTTP probe logic
│   └── scheduler.go     # Per-target goroutine management with mutex-protected state
├── config/              # Environment-based configuration
├── db/
│   ├── db.go            # Init, migrations
│   ├── targets.go       # Target CRUD
│   └── checks.go        # Check result storage and queries
└── events/              # In-process pub/sub event bus
    ├── bus.go
    └── event.go
```

## API

| Method   | Endpoint                            | Description                           |
| -------- | ----------------------------------- | ------------------------------------- |
| `GET`    | `/api/targets`                      | List all targets                      |
| `GET`    | `/api/targets/:id`                  | Get a single target                   |
| `POST`   | `/api/targets`                      | Create a target                       |
| `PUT`    | `/api/targets/:id`                  | Update a target                       |
| `DELETE` | `/api/targets/:id`                  | Delete a target                       |
| `GET`    | `/api/targets/:id/history?limit=50` | Get check history                     |
| `GET`    | `/api/events`                       | Subscribe to real-time event stream   |

### Create a target

```bash
curl -X POST http://localhost:8080/api/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Website",
    "url": "https://example.com",
    "method": "GET",
    "interval_seconds": 30,
    "timeout_seconds": 5,
    "expected_status": 200,
    "active": true
  }'
```
### Subscribe to the event stream

```bash
curl -N http://localhost:8080/api/events
```

The endpoint emits Server-Sent Events in the form:

```
data: {"type":"check_complete","target_id":1,"at":"2026-04-17T15:58:39Z","success":true,"status_code":200,"response_time_ms":323}

data: {"type":"state_change","target_id":1,"at":"2026-04-17T15:58:44Z","from":"up","to":"down"}
```

Event types:

- `check_complete` — emitted after every check, includes `status_code`, `response_time_ms`, `success`, and optional `error_message`
- `state_change` — emitted when a target transitions between `up` and `down`, includes `from` and `to`

Multiple clients can subscribe concurrently; each receives the full event stream independently.

## Quick Start

### Prerequisites

- Go 1.22+
- GCC and SQLite (`libsqlite3-dev` on Debian/Ubuntu, `sqlite` on Arch)

### Run locally

```bash
# Create a .env file
cat > .env << EOF
SERVER_PORT=8080
DB_PATH=./servicepatrol.db
CORS_ORIGIN=*
EOF

# Build and run
CGO_ENABLED=1 go build -o servicepatrol ./cmd/server/main.go
./servicepatrol
```

### Deploy to K3s

```bash
# Copy and configure environment
cp .env.deploy.example .env.deploy
# Edit .env.deploy with your values

# First-time deployment
./setup.sh

# Subsequent updates
./update.sh
```

## Configuration

| Variable           | Default                  | Description                                   |
| ------------------ | ------------------------ | --------------------------------------------- |
| `SERVER_PORT`      | `8080`                   | HTTP server port                              |
| `DB_PATH`          | `/data/servicepatrol.db` | SQLite database path                          |
| `ALERTMANAGER_URL` | _(empty)_                | Alertmanager endpoint; leave empty to disable |
| `CORS_ORIGIN`      | `*`                      | Allowed CORS origin                           |

## Design Decisions

- **No auth** — ServicePatrol is designed to run behind a private network or be called server-side from an authenticated dashboard, not exposed directly to the internet
- **SQLite over PostgreSQL** — single-file database with zero config, ideal for homelab deployments with low write volume
- **Per-target goroutines** — each target gets its own goroutine with independent scheduling, making it easy to add, remove, or reconfigure targets at runtime without affecting others
- **Optional Alertmanager** — runs standalone without alerting; toggle on by setting the URL
- **Frontend-agnostic** — pure REST API with no embedded UI, designed to integrate with any frontend
- **Pub/sub event bus** — an in-process bus fans out check results and state transitions to multiple subscribers (SSE clients, Alertmanager notifier), keeping the scheduler decoupled from consumers
- **Non-blocking event publishing** — if a subscriber's buffer fills, events are dropped for that subscriber rather than blocking the scheduler, preserving liveness under slow consumers

