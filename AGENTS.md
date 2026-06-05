# AGENTS — AI Coding Agent Guidance

Purpose

- Provide minimal, actionable guidance so AI coding agents can be productive in this repository.

Quick commands

- Build: `go build ./...`
- Run API locally: `go run ./cmd/api`
- Dev: `air` (live-reload development runner)
- Run migrations: `go run migration.go`
- Tests: `go test ./...`

Project layout (high level)

- `cmd/` — application entrypoints (see [cmd/api/main.go](cmd/api/main.go)).
- `internal/` — service domains; each domain follows the pattern: `handlers/`, `service/`, `repository/` (GORM implementations in `repository/gorm.go`).
- `database/connection.go` — centralized DB connection helpers.
- `models/` — shared models.
- `migration.go` — DB migration runner.
- `go.mod` — dependency manifest (use `go` toolchain from `go.mod`).

Conventions and patterns

- Routing and HTTP: handlers and route wiring live under `internal/*/handlers`.
- Business logic lives in `internal/*/service` and is invoked by handlers.
- Persistence: repository implementations use GORM; see `internal/*/repository/gorm.go` for examples.
- Keep changes focused: prefer small PRs that modify one domain (artist, lyrics, user) at a time.

Agent behaviour guidelines

- Link, don't embed: reference workspace files rather than copying long docs.
- Safety-first edits: run `go test ./...` after changes; prefer non-breaking refactors.
- Use existing patterns: follow the handler→service→repository flow seen in `internal/*`.
- When editing DB code, check `database/connection.go` and `migration.go` for initialization patterns.

Useful files to inspect

- [README.md](README.md)
- [go.mod](go.mod)
- [migration.go](migration.go)
- [cmd/api/main.go](cmd/api/main.go)
- [database/connection.go](database/connection.go)
- Examples: [internal/artist/repository/gorm.go](internal/artist/repository/gorm.go)

Suggested next customizations

- Create a `run-and-test` skill that runs `go test ./...` and `go run ./cmd/api` in a disposable environment.
- Add a `code-review` prompt that enforces small, focused PRs and checks for database migration impact.

Feedback

- If you'd like, I can add a `.github/copilot-instructions.md` variant, or split instructions per subsystem.

**Response Format**

- **Envelope**: All API responses use a standardized JSON envelope:

- **Envelope**: All API responses use a standardized JSON envelope:

  ```json
  {
  	"code": 0,        // 0 for success, negative for errors
  	"message": "...", // human-readable message
  	"data": { ... }   // payload or null on error
  }
  ```

- **Status Codes**: The HTTP status code is always 200 for API responses; the `code` field inside the envelope indicates success or failure. Project conventions:
  - **0**: success (example message: "success")
  - **-1**: generic failure ("failed")
  - **-2**: bad request ("bad request")
  - **-3**: not found ("not found")
  - **-4**: internal server error ("internal server error")

- **Examples**:

  Success:

  ```json
  {
    "code": 0,
    "message": "success",
    "data": { "id": 123, "name": "Example" }
  }
  ```

  Error (username not found):

  ```json
  {
    "code": -100,
    "message": "username doesn't exist",
    "data": null
  }
  ```

Notes

- Default messages are provided by `internal/util/response.go` but handlers and services may supply custom messages when returning non-zero codes.
