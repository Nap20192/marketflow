# MarketFlow

MarketFlow is a modular, production-oriented service for collecting, aggregating, caching, and storing market data. Built in Go with PostgreSQL and Redis, it follows clean-architecture principles and idiomatic Go concurrency patterns so the codebase is easy to extend and operate.

This README is a short, practical guide to help you get the project running locally, understand the repo layout, and contribute.

## Highlights

- Aggregates market data from multiple exchange generators
- Persists reliable data in PostgreSQL for analytics and historical queries
- Uses Redis for fast, cache-friendly access patterns
- Designed with adapters, ports, and services (clean architecture)
- Concurrency helpers (worker pools, fan-in/fan-out) for high-throughput processing

## Quick Start

The easiest way to run everything (Postgres, Redis, and the exchange generator images) is with Docker Compose.

1. Clone the repository:

   ```sh
   git clone https://github.com/Nap20192/marketflow.git
   cd marketflow
   ```

2. Bring up the stack (builds the Go service image and starts infra):

   ```sh
   docker-compose up --build
   ```

3. Check logs or application status:

   ```sh
   make logs
   make status
   ```

If you prefer to run the Go binary directly during development, ensure Postgres and Redis are available and then:

```sh
go mod download
go run ./cmd/main.go --port=8080
```

Or use the convenience Make target:

```sh
make run
```

## Prerequisites

- Go 1.20+
- Docker & Docker Compose (for the quick-start)
- (Optional) psql client to inspect the database

## Project Layout

Key directories and their purpose:

- `cmd/` — application entry point (`cmd/main.go`)
- `generator/` — bundled exchange generator images (Docker .tar files)
- `infrastucture/` — integrations with external systems (Postgres, Redis)
  - `postgres/` — DB connection, queries and fallback logic
  - `redis/` — cache layer and querier
- `internal/` — application wiring, adapters and HTTP server
  - `adapters/primary/exchanger/` — external data sources
  - `adapters/primary/ui/` — HTTP server, routes and handlers
  - `adapters/secondary/` — cache/storage adapters
  - `core/` — domain interfaces (ports), models and services
- `pkg/` — shared utilities (concurrency, error groups, logging)

## Configuration

Configuration is driven by Go structs and environment variables loaded at startup. See:

- `internal/config.go` — top-level application config
- `infrastucture/postgres/config.go` — Postgres connection options
- `infrastucture/redis/config.go` — Redis connection options

Use environment variables or a `.env` file when running with Docker Compose to override defaults.

## Makefile Targets

Useful targets (see `Makefile` for exact behavior):

- `make up` — Loads exchange images and starts services (detached)
- `make down` — Stops and removes containers and volumes
- `make load-exchanges` — Loads the bundled exchange Docker images from `generator/*.tar`
- `make run` — Runs the Go application locally (`go run cmd/main.go`)
- `make rebuild` — Rebuilds images and restarts the stack
- `make logs` — Tail Docker Compose logs
- `make status` — Show Docker Compose service status

## Testing

Run unit tests:

```sh
go test ./...
```

Some packages have focused tests (adapters, infra). Integration tests may require Postgres/Redis.

## Contributing & Development Notes

- Add new exchange sources under `internal/adapters/primary/exchanger` and wire them into the aggregator service.
- Keep business logic in `core/service` and I/O in adapters.
- Follow Go idioms: explicit error returns, `context.Context` usage, and small packages.

Developer checklist:

- Run `go mod tidy` after adding dependencies
- Add unit tests for new behavior
- Run `go vet` and `golangci-lint` (if used) before creating PRs

## Runtime & Observability

- Logging: `pkg/logger` (structured, contextual logs)
- Troubleshooting: `make logs`, `make status`

## License

This repository does not include an explicit license file. Contact the maintainers if you plan to use or contribute.

---
