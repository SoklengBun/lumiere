# AGENTS — AI Coding Agent Guidance

Use this repository as a Go backend service built with Echo, GORM, and PostgreSQL. Keep changes focused, follow the existing layered structure, and prefer small, reviewable edits.

## Quick commands

- Build: `go build ./...`
- Run API locally: `go run ./cmd/server`
- Apply schema changes: `go run migration.go`
- Run tests: `go test ./...`

## Runtime entrypoints

- App wiring: [internal/app/app.go](internal/app/app.go)
- Server startup: [cmd/server/main.go](cmd/server/main.go)
- Vercel-style entrypoint: [api/index.go](api/index.go)

## Architecture conventions

- The app is split by domain into handler, service, repository, and model folders under [internal](internal).
- HTTP route registration happens in handlers, and service-layer business logic stays out of handlers.
- Persistence is handled with GORM repositories; inspect existing repo implementations for the expected pattern.
- Shared response behavior lives in [internal/util/response.go](internal/util/response.go).

## When making changes

- Prefer one domain at a time and keep edits scoped.
- Follow the existing flow: handler → service → repository.
- If a change touches persistence, inspect [internal/database/connection.go](internal/database/connection.go) and [migration.go](migration.go) before editing.
- Redis is optional. Treat cache failures as non-fatal; the app should continue serving through PostgreSQL.
- Before claiming a change is complete, rerun `go test ./...` and confirm the relevant behavior with fresh evidence.

## Useful references

- [README.md](README.md)
- [SYSTEM_DESIGN.md](SYSTEM_DESIGN.md)
- [go.mod](go.mod)
- [internal/config/config.go](internal/config/config.go)
- Domain example: [internal/artist/repository/gorm.go](internal/artist/repository/gorm.go)

## Agent behavior

- Link to existing workspace docs instead of copying long explanations into new edits.
- Prefer non-breaking refactors and minimal diffs.
- When the change affects schema or environment assumptions, update the related docs and migration behavior together.
