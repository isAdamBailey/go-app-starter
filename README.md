# go-app-starter

[![Backend CI](https://github.com/isAdamBailey/go-app-starter/actions/workflows/backend-ci.yml/badge.svg)](https://github.com/isAdamBailey/go-app-starter/actions/workflows/backend-ci.yml)
[![Frontend CI](https://github.com/isAdamBailey/go-app-starter/actions/workflows/frontend-ci.yml/badge.svg)](https://github.com/isAdamBailey/go-app-starter/actions/workflows/frontend-ci.yml)

A starter template for small, personal-scale full-stack apps: Go API +
Postgres + Nuxt SPA, with passwordless magic-link auth, email, and Docker
already wired up.

Use this repo as a GitHub template ("Use this template" button) or clone it
directly to start a new app.

## Stack

- **Backend**: Go, [chi](https://github.com/go-chi/chi) router, pgx v5 /
  pgxpool, [sqlc](https://sqlc.dev)-generated queries, golang-migrate with
  embedded migrations
- **Frontend**: Nuxt 4 SPA (`ssr: false`), Tailwind v4, Pinia
- **Database**: PostgreSQL 16
- **Local dev**: Docker Compose (Postgres + Mailpit + backend + frontend)
- **Production**: Laravel Forge on a VPS — see `docs/DEPLOY.md`

Only need the Go backend (e.g. to drop into an existing Nuxt monorepo)? See
`docs/MONOREPO.md`.

## Getting started

```sh
git clone <this-repo> my-new-app
cd my-new-app
./scripts/new-app.sh github.com/you/my-new-app "My New App" my_new_app
cp .env.example .env   # then set COOKIE_SIGNING_SECRET and ALLOWED_EMAILS
docker compose up --build
```

- Backend: http://localhost:8080 (`/healthz`)
- Frontend: http://localhost:3000
- Mailpit (catches magic-link emails in dev): http://localhost:8025
- Postgres: localhost:5432

Sign in at `/login` with an email from `ALLOWED_EMAILS`; the magic link
lands in Mailpit.

## What's here

### Auth (`backend/internal/auth`, `internal/users`)

Passwordless magic-link login: request a link, click it, get a signed
HttpOnly session cookie plus a CSRF cookie (double-submit pattern). Sessions
are server-side rows (`sessions` table), so logout/revocation is immediate —
no JWTs to invalidate. Access is gated by an email allowlist
(`ALLOWED_EMAILS`), synced into the `allowed_users` table on every boot.

### Email (`backend/internal/mailer`)

One `Mailer` interface, three implementations selected by `EMAIL_PROVIDER`:
`smtp` (Mailpit locally), `ses` (AWS SES over SMTP), `resend` (Resend HTTP
API).

### HTTP API (`backend/internal/httpapi`)

chi router with: security headers (`security.go`), per-IP rate limiting via
[go-chi/httprate](https://github.com/go-chi/httprate) on the auth endpoints,
CSRF enforcement (`requireCSRF` middleware) on state-changing routes, request
body size limits, and a trusted-proxy-aware real-IP resolver (`realip.go`) —
see the comment there for why this doesn't use chi's deprecated
`middleware.RealIP`.

### Database (`backend/internal/db`, `migrations/`, `db/queries/`)

sqlc generates typed Go from `db/queries/*.sql` against the schema in
`migrations/`. Migrations are embedded into the binary and applied
automatically on startup (`cmd/server`) or standalone (`cmd/migrate`).
Regenerate after changing a query:

```sh
cd backend && go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate
```

### Testing

Backend tests use hand-written in-memory fakes (`fakes_test.go` in each
package) implementing that package's `Querier`/`Repository` interfaces — no
real database in unit tests. `internal/httpapi` tests spin up a `chi.Router`
via `httptest` and exercise the full magic-link → session → CSRF flow.

```sh
cd backend
go test ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## Security & operability choices worth knowing about

- **Structured logging** via `log/slog`.
- **Graceful shutdown** and `http.Server` timeouts (`ReadHeaderTimeout` etc.)
  in `cmd/server`, so a slow client can't hold a connection open forever.
- **`go-chi/httprate`** for per-IP rate limiting on auth endpoints — a
  maintained library rather than a hand-rolled limiter.
- **A trusted-proxy-aware real-IP middleware** (`realip.go`) instead of
  chi's deprecated `middleware.RealIP`, which trusts `X-Forwarded-For`
  unconditionally and is spoofable by any client — a real concern for
  IP-based rate limiting specifically.
- **Security headers middleware** (`X-Content-Type-Options`,
  `X-Frame-Options`, CSP, etc.) and a **request body size cap** on JSON
  endpoints.
- **Distroless, non-root Docker runtime image** for the backend — no shell,
  no package manager, minimal attack surface.
- **`gosec` and `govulncheck`** wired into linting/CI.
- **`nuxt-security`** module on the frontend for response headers.
- A minimum-length check on `COOKIE_SIGNING_SECRET` at config-load time.

## Deploying

See `docs/DEPLOY.md` for the full Forge + VPS setup this template is built
for (nginx reverse proxy, PM2 for Nuxt, a Go daemon, environment variable
reference, troubleshooting table).
