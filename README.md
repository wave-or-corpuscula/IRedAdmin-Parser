# IRedAdmin Parser

Scrape, sync, and browse iRedAdmin mail server data — a dual-component toolkit for managing iRedMail infrastructure at scale.

[![Go Version](https://img.shields.io/badge/Go-1.26.3-00ADD8?logo=go)](./iredparser/)
[![Python Version](https://img.shields.io/badge/Python-3.13+-3776AB?logo=python)](.)
[![SQLite](https://img.shields.io/badge/SQLite-modernc_/_stdlib-003B57?logo=sqlite)](./iredparser/internal/database/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## Overview

iRedAdmin Parser connects to iRedAdmin — the web admin panel for iRedMail servers — scrapes domain and mailbox metadata, and stores it locally. Two interfaces, shared SQLite data:

- **[`iredparser/`](iredparser/) — Go CLI**: Concurrent scraping engine. Authenticates against iRedAdmin, parses HTML into structured data, upserts into SQLite.
- **[`app/`](app/) — Python TUI**: Terminal browser built with Textual. Search, filter, sort parsed data; trigger syncs and password changes interactively.

---

## Go CLI (`iredparser/`)

The core scraping engine written entirely in Go.

### Architecture

Clean-ish hexagonal layout with explicit package boundaries:

```
iredparser/
  cmd/parser-cli/   # Entry point
  common/           # Shared types (ServerConfig)
  internal/
    controller/     # Cobra command handlers, dependency injection root
    database/       # SQLite persistence (sqlx + modernc.org/sqlite)
    parser/
      client/       # HTTP client with cookie-jar session management
      domain/       # Domain list scraper
      mailbox/      # Concurrent mailbox scraper (worker pool)
    services/
      auth_service/       # iRedAdmin authentication
      password_service/   # Remote password change
    sync/           # Orchestration: scrape → persist pipeline
  pkg/
    errors/         # Typed sentinel error hierarchy
    utils/          # Memory-size string parsing, CSRF extraction
  testing/          # Shared test helpers
```

### Tech & Patterns

**HTTP Client** (`internal/parser/client/`)

- `http.Client` with custom `*cookiejar.Jar` for session cookie persistence across requests
- `TLSClientConfig{InsecureSkipVerify: true}` for self-signed certs common in mail server deployments
- Connection pooling: 100 idle conns, 10 per host, 90s idle timeout
- User-Agent masquerading (Chrome 149) to bypass naive bot detection
- Structured error handling: each HTTP failure maps to a typed sentinel error (`ErrPostRequestCreation`, `ErrUnexpectedStatusCode`, `ErrInternalServerError`)
- Session cookie extraction from cookie jar after successful auth

**Concurrent HTML Scraping** (`internal/parser/mailbox/`)

- Pagination-aware: discovers page count from `.pages` nav element, fans out page fetches
- Worker pool: 30 goroutines reading from a buffered `chan string` (job URLs), writing to `chan []*parser.Mailbox`
- `sync.WaitGroup` via `wg.Go()` (Go 1.26 ergonomics) for coordinated shutdown
- `goquery` (jQuery-style DOM traversal) for HTML parsing — extracts mailboxes from `<tbody> <tr>` rows on each page
- Graceful error tolerance: per-page parse failures are skipped (logged in production), remaining results preserved

**Database** (`internal/database/`)

- `jmoiron/sqlx` with `modernc.org/sqlite` — pure Go SQLite, zero CGO dependency
- `ON CONFLICT ... DO UPDATE ... RETURNING` — atomic upsert that returns the row ID in one round trip, eliminating separate SELECT-then-INSERT races
- Transactional batch operations (`Beginx`/`Commit`/`Rollback`) for `UpsertDomainMany` / `UpsertMailboxMany` — full rollback on any failure
- Embedding via struct composition: `DomainModel` embeds `parser.Domain`, `MailboxModel` embeds `parser.Mailbox` — DB models carry domain types without copy/paste mapping
- Auto-schema initialization on connect (`CREATE TABLE IF NOT EXISTS`) with foreign key enforcement

**CLI Framework** (`internal/controller/`)

- `spf13/cobra` with subcommands: `auth-check`, `sync`, `change-password`
- Middleware pattern via `PersistentPreRunE`: JSON config parsing → auth → inject config into controller state — runs before every command
- Dependency injection in `NewCLIController(client, storage, authService, syncService, passwordService, writer)` — no global state, testable in isolation
- Interface segregation: `AuthChecker`, `SyncService`, `PasswordChanger`, `Storage` — each interface is one method, one responsibility

**Error Handling** (`pkg/errors/`)

- Custom `IRedError` type with `ErrType` (authentication, HTTP, parsing, CLI) and `ErrCode` (numeric, grouped by domain)
- Sentinel errors with `errors.Is()` matching on (Type, Code) pair, not just message string
- `IsType()` helper for type-level matching (e.g., `IsType(err, ErrTypeAuthentication)`)
- `IRedMultiError` implements `Unwrap() []error` for multi-error aggregation
- Consistent error wrapping with `%w` through the call chain — no information loss

### CLI Usage

```bash
# Build
cd iredparser && go build -o bin/iredparser ./cmd/parser-cli/main.go

# Auth check
./iredparser -c '{"server":"mail.example.com","login":"admin@example.com","password":"secret"}'

# Sync mailbox data
./iredparser -c '{"server":"mail.example.com","login":"admin@example.com","password":"secret"}' sync

# Change password
./iredparser -c '{"server":"mail.example.com","login":"admin@example.com","password":"secret"}' change-password

# With config file
./iredparser -c "$(cat config.json)" sync
```

---

## Python TUI (`app/`)

Terminal interface built with [Textual](https://textual.textualize.io/) for browsing and managing parsed mail server data.

### Screens

| Screen | Purpose |
|---|---|
| **Main Menu** | Navigation hub |
| **Search** | Table view with live filtering (server, admin status, ban status, quota), column sorting, full-text search |
| **Config** | Add / remove / test server connections |
| **Sync** | Trigger sync per server or all at once |
| **Progress** | Real-time password change progress with per-mailbox status |

### Patterns

- `@dataclass` + Textual `reactive` for automatic UI refresh on filter/state changes
- Repository pattern on the database layer (`ServerRepository`, `DomainRepository`, `MailboxRepository`)
- `@work` decorator for background async sync operations without UI freeze
- Async/await throughout — Textual's async event loop drives all I/O

---

## Tech Stack

### Go

| Dependency | Purpose |
|---|---|
| **Go 1.26.3** | Generics, `t.Context()`, `wg.Go()` |
| **[cobra](https://github.com/spf13/cobra)** v1.10 | CLI framework, subcommands, middleware |
| **[goquery](https://github.com/PuerkitoBio/goquery)** v1.12 | jQuery-style HTML parsing |
| **[sqlx](https://github.com/jmoiron/sqlx)** v1.4 | `database/sql` with named queries + StructScan |
| **[modernc.org/sqlite](https://modernc.org/sqlite)** v1.52 | Pure Go SQLite driver (no CGO) |
| **[testify](https://github.com/stretchr/testify)** v1.11 | Test assertions + suite support |
| stdlib `net/http/cookiejar` | Session cookie management |
| stdlib `crypto/tls` | TLS with insecure-skip for self-signed certs |

### Python

| Dependency | Purpose |
|---|---|
| **[Textual](https://textual.textualize.io/)** | Terminal UI framework |
| **stdlib sqlite3** | SQLite access |
| **stdlib asyncio** | Async concurrency |

### Infrastructure

| Tool | Purpose |
|---|---|
| **SQLite** | Embedded data store — shared between Go CLI and Python TUI |
| **ruff** | Python linter + formatter |
| **pyright** | Python static type checker |

---

## Testing

### Go (`iredparser/`)

- **Unit tests**: in-memory SQLite (`:memory:`) for fast database tests — schema created, queries executed, connection dropped — no filesystem noise
- **Integration tests**: real iRedAdmin server via `.test.creds.json` credential file. Tests authenticate, scrape, and verify against a live or staging mail server
- **HTTP-level tests**: structured HTTP error handling tested with expected status code → sentinel error mappings
- Coverage includes: client auth, domain parser, mailbox parser, domain sync, mailbox sync, utils, CLI controller

### Python (`app/`)

- **pytest** with Textual testing utilities
- **Fixtures**: in-memory database sessions, config storage, mock HTTP responses
- **Screens**: widget state tests for search, config, sync flows

```bash
# Run Go tests (unit only, no server required)
cd iredparser && go test ./...

# Run Go integration tests (requires .test.creds.json)
cd iredparser && go test -tags=integration ./...

# Run Python tests
pytest
```

---

## Getting Started

### Prerequisites

- Go 1.26+
- Python 3.13+
- Access to an iRedAdmin web interface

### Quick Start

```bash
# Clone
git clone <repo-url> && cd IRedAdmin-Parser

# Build Go CLI
cd iredparser && go build -o bin/iredparser ./cmd/parser-cli/main.go

# Set up Python TUI
python3 -m venv .venv && source .venv/bin/activate
pip install textual

# Create a credentials file
cp .test.creds.json.dummy config.json
# Edit config.json with your server details

# Sync data
./iredparser/bin/iredparser -c "$(cat config.json)" sync

# Browse
source .venv/bin/activate && python run.py
```

---

## Project Layout

```
IRedAdmin-Parser/
  iredparser/               # Go CLI — scraping engine
    cmd/parser-cli/main.go  # Entry point
    common/                 # Shared config types
    internal/
      controller/           # Cobra command handlers + DI
      database/             # SQLite persistence layer
      parser/
        client/             # HTTP client with cookie jar
        domain/             # Domain scraper
        mailbox/            # Concurrent mailbox scraper
      services/
        auth_service/       # Auth against iRedAdmin
        password_service/   # Password change remote call
      sync/                 # Scrape → persist orchestration
    pkg/
      errors/               # Typed error hierarchy
      utils/                # Memory parsing, CSRF extraction
  app/                      # Python TUI — data browser
    tui/screens/            # Textual screen definitions
    tui/widgets/            # Reusable UI components
    backend/                # HTTP client for external calls
    database/               # SQLite repository layer
    services/               # Business logic (sync, config, passwords)
    storage/                # Config persistence
  data/                     # SQLite database location (gitignored)
```

---

## Design Decisions

- **CGO-free SQLite** — `modernc.org/sqlite` translates SQLite to Go assembly rather than linking libsqlite. Zero C toolchain dependency, trivial cross-compilation (a problem on mail servers that often run older Linux distros).
- **Worker pool over goroutine-per-page** — page count is unknown until the first request completes. A fixed-size pool (30 workers) prevents unbounded goroutine growth while keeping throughput high for 100+ page servers.
- **UPSERT with RETURNING** — avoids the SELECT-then-INSERT race in concurrent syncs. If two syncs hit the same domain simultaneously, the second upsert won't fail — it updates in place and returns the existing row ID.
- **Struct embedding for models** — `DomainModel` embeds `parser.Domain` so adding a scraped field requires exactly one change (add field to model + add column to schema). No manual mapping layer.

---

## License

MIT