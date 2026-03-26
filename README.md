# ServicePatrol

Lightweight HTTP health monitoring service written in Go. Designed for self-hosted environments and homelabs.

ServicePatrol monitors your services by probing HTTP endpoints at configurable intervals, stores check history in SQLite, and optionally fires alerts through Alertmanager when a target goes down or recovers.

## Features

- **Per-target scheduling** — each monitored service runs in its own goroutine with independent intervals and timeouts
- **SQLite persistence** — zero-dependency storage with automatic schema migrations
- **Alertmanager integration** — optional firing and auto-resolving alerts on state changes
- **REST API** — full CRUD for targets and paginated check history, compatible with any frontend
- **Minimal footprint** — single binary, ~5MB memory at runtime

## Architecture

```
┌──────────────┐      ┌─────────────┐      ┌──────────────────┐
│   REST API   │◄────►│  SQLite DB  │◄────►│    Scheduler     │
│  (handlers)  │      │  (targets,  │      │  (per-target     │
│              │      │   checks)   │      │   goroutines)    │
└──────────────┘      └─────────────┘      └───────┬──────────┘
                                                   │
                                           ┌───────▼──────────┐
                                           │   HTTP Checker   │
                                           │  (probe targets) │
                                           └───────┬──────────┘
                                                   │
                                           ┌───────▼──────────┐
                                           │  Alertmanager    │
                                           │  (optional)      │
                                           └──────────────────┘
```

```
internal/
├── api/
│   ├── handlers/       # REST endpoints for targets and history
│   └── middleware/      # CORS
├── alertmanager/        # Alertmanager client (fire/resolve)
├── checker/
│   ├── http.go          # HTTP probe logic
│   └── scheduler.go     # Per-target goroutine management
├── config/              # Environment-based configuration
└── db/
    ├── db.go            # Init, migrations
    ├── targets.go       # Target CRUD
    └── checks.go        # Check result storage and queries
```

## API

| Method   | Endpoint                            | Description         |
| -------- | ----------------------------------- | ------------------- |
| `GET`    | `/api/targets`                      | List all targets    |
| `GET`    | `/api/targets/:id`                  | Get a single target |
| `POST`   | `/api/targets`                      | Create a target     |
| `PUT`    | `/api/targets/:id`                  | Update a target     |
| `DELETE` | `/api/targets/:id`                  | Delete a target     |
| `GET`    | `/api/targets/:id/history?limit=50` | Get check history   |

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

