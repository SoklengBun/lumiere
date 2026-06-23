# Lumiere System Design

## Purpose

This document describes the current backend system as implemented in the repository today. It is intended as a quick review artifact for understanding the architecture, data model, API surface, and known gaps relative to the product plan.

## System Scope

Lumiere is a Go backend for a music and lyrics catalog with:

- user registration and login
- artist management
- lyrics/song management
- playlist management
- support for song-to-song cover relationships
- support for artist CV relationships, so a character artist can reference the real performer behind that artist

## Runtime Architecture

The application uses a layered structure per domain:

- handler: HTTP request parsing and response shaping
- service: business logic orchestration
- repository: persistence with GORM
- model: domain and database structs

Main runtime flow:

1. Echo server starts from `cmd/server/main.go`.
2. `internal/app/app.go` loads env config and opens the PostgreSQL connection.
3. Domain repositories, services, and handlers are wired together.
4. Routes are mounted under `/api`.
5. All endpoints return the shared JSON envelope from `internal/util/response.go`.

There is also a Vercel-style handler entrypoint in `api/index.go` that reuses the same Echo app.

## High-Level Components

### App Layer

- `internal/app/app.go`
- creates the Echo instance
- enables logger, panic recovery, and permissive CORS
- wires all domain modules

### Config Layer

- `internal/config/config.go`
- reads required env vars for database connection and JWT signing

Required environment variables:

- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASS`
- `JWT_SECRET`

Optional environment variables:

- `DB_SSLMODE`
- `APP_HOST`
- `APP_PORT`
- `PORT`
- `DB_MAX_OPEN_CONNS`
- `DB_MAX_IDLE_CONNS`
- `DB_CONN_MAX_LIFETIME_MIN`

### Database Layer

- `internal/database/connection.go`
- PostgreSQL via GORM
- connection pool configured from env

### Migration Layer

- `migration.go`
- runs GORM `AutoMigrate`
- includes a compatibility migration to preserve legacy string lyrics IDs as `videoId` values while restoring numeric primary keys

## Domain Model

### User

Source: `internal/models/user.model.go`

Fields:

- `id`
- `username`
- `password`
- `name`
- `role`

Notes:

- users are returned through a sanitized public shape where needed
- roles exist but most current access control is ownership-based rather than role-based

### Artist

Source: `internal/artist/model.go`

Fields:

- `id`
- `name`
- `normalizedName`
- `cv []Artist`

Notes:

- `cv` is a self-referential many-to-many relation through `artist_cvs`
- this supports cases like `A (cv B)`
- direct artist and CV artist are intentionally asymmetric in search behavior

### Lyrics

Source: `internal/lyrics/model.go`

Fields:

- `id`
- `videoId string`
- `title string`
- `altTitles []string`
- `artists []Artist`
- `contents []LyricContent`
- `covers []LyricCover`
- `createdById`

Notes:

- lyrics entries represent songs
- `id` is the stable numeric primary key
- `videoId` stores the editable YouTube video ID
- product-wise, a cover is intended to be another performance of the same song, not a full second song object
- each cover should only need its own YouTube ID and performer list

### LyricCover

Source: `internal/lyrics/model.go`

Fields:

- `id string`
- `artists []Artist`

Notes:

- `id` should be the cover video's YouTube ID
- `artists` should be the performers for that cover instance
- this matches the product requirement where the base song owns the canonical title and lyric content, while each cover only stores performer-specific data
- the API accepts cover objects through `covers`

### LyricContent

Fields:

- `lyricsId`
- `kind`
- `content`

Notes:

- supports multi-language or multi-format lyric storage such as japanese, romaji, or english

### Playlist

Source: `internal/playlist/model.go`

Fields:

- `id`
- `name`
- `description`
- `isPublic`
- `createdById`
- `items []PlaylistItem`

### PlaylistItem

Fields:

- `id`
- `playlistId`
- `lyricsId`
- `position`
- `note`
- `defaultCoverId`
- `lyrics`

Notes:

- playlist order is explicit via `position`
- playlist items resolve to lyrics records
- playlist items can pin a cover via `defaultCoverId`

## Relationship Summary

- one user creates many lyrics
- one user creates many playlists
- many lyrics belong to many artists through `lyrics_artists`
- one lyrics has many titles
- one lyrics has many content variants
- one lyrics should have many cover entries
- one cover entry should have many artists through a dedicated cover-artist join table
- one artist can reference many CV artists through `artist_cvs`

## Search Semantics

### Artist Search

Current artist search supports:

- direct name match on artist name or normalized name
- CV traversal match

Intended result shape:

- searching `A` returns artist `A`
- searching `B` returns artist `B` and artists whose `cv` contains `B`

This gives the desired behavior:

- `A (cv B)` is found when searching for `B`
- searching for `A` does not surface `B` just because `B` is the CV

### Lyrics Search

Lyrics search is exposed at `/api/lyrics/search`, but detailed ranking and matching rules should be treated as implementation-level behavior in the service/repository rather than a stable external contract at this stage.

### Playlist Search

Playlist search is exposed at `/api/playlist/search` and is intended for public playlist discovery.

## Authentication and Authorization

Authentication is token-based:

- register and login return a token
- protected endpoints accept the `Authorization` header
- `Bearer <token>` format is supported

Current authorization rules:

- private playlists can only be viewed by their owner
- playlist mutation requires ownership
- lyrics edit requires ownership
- lyrics add can attach `createdById` if a valid token is present
- artist endpoints are currently lightweight and do not yet enforce the same ownership rules

## API Surface

All routes are mounted under `/api`.

### User

- `POST /api/user/register`
- `POST /api/user/login`
- `GET /api/user/quick-login`

### Artist

- `GET /api/artist/:id`
- `GET /api/artist/list`
- `GET /api/artist/search?q=...`
- `POST /api/artist/add`
- `PUT /api/artist/:id/cv`

### Lyrics

- `GET /api/lyrics/search?q=...`
- `GET /api/lyrics/:id`
- `GET /api/lyrics/list`
- `GET /api/lyrics/mine`
- `POST /api/lyrics/add`
- `PUT /api/lyrics/:id`

### Playlist

- `GET /api/playlist/search?q=...`
- `GET /api/playlist/:id`
- `GET /api/playlist/list`
- `GET /api/playlist/mine`
- `POST /api/playlist/add`
- `PUT /api/playlist/:id`
- `DELETE /api/playlist/:id`
- `POST /api/playlist/:id/items`
- `PUT /api/playlist/:id/items/reorder`
- `DELETE /api/playlist/:id/items/:itemId`

## Response Contract

All API responses use a standard envelope:

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

Common codes:

- `0` success
- `-1` generic failure
- `-2` bad request
- `-3` not found
- `-4` internal server error
- `-5` unauthenticated

HTTP status is generally kept at `200`, with the envelope code carrying application outcome.

## Current Behavior Mapped To Product Plan

### Already Supported

- users can register and login
- songs can be created with editable `videoId` values
- songs can have multiple titles
- songs can have multiple artists
- songs can have multiple lyric content versions
- songs can store covers as lightweight entries with YouTube ID plus artists
- playlists support public/private visibility
- home-oriented compact playlist shaping exists in playlist handlers
- artist CV support now exists in the data model and artist search path

### Partially Supported

- cover display can model `A (cv B)`, but UI formatting still needs to be implemented wherever songs/covers are rendered
- search exists per resource, but cross-resource home search from the plan is not yet represented as a single aggregated endpoint
- playlist list behavior trims to five songs in handler shaping for home-mode use, but random public-five selection is not yet documented as a guaranteed backend behavior

### Not Yet Explicitly Implemented In This Review

- a dedicated homepage aggregation endpoint returning random playlists plus compact songs
- a dedicated song detail response optimized exactly to the plan’s display contract
- richer artist aliases or multilingual artist naming
- stronger role-based admin flows
- explicit audit/history/versioning for lyrics edits

## Known Design Notes

### CV Modeling Choice

The current CV design stores CV at the artist level, not at the cover edge level.

Implication:

- if artist `A` has CV `B`, that relationship applies consistently wherever `A` is used
- this is good for reuse and searchability
- if the product later needs per-song or time-varying CV attribution, the model may need a more specific join entity rather than a plain self-referential artist relation

### Cover Modeling Choice

The intended cover relation is song-to-cover-entry, not song-to-song.

Implication:

- the base song keeps the canonical title and lyric content
- each cover entry stores only its own YouTube ID and artists
- this better matches the plan's shape: `covers -> id + artists`

## Suggested Next Steps

1. Add a dedicated review endpoint or doc section for request and response examples per route.
2. Add a homepage aggregation endpoint that matches the product plan exactly.
3. Add explicit response DTOs for song detail and cover display formatting.
4. Decide whether CV should remain artist-global or become cover-specific if future use cases require that precision.
