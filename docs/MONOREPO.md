# Using this in an existing Nuxt monorepo

This template ships as `backend/` + `frontend/` because that matches how it
gets deployed (see `docs/DEPLOY.md`). If you already have a Nuxt app and just
want its Go API — magic-link auth, sessions, Postgres, email, Docker — drop
in `backend/` and skip `frontend/` entirely.

## Steps

1. Copy `backend/`, `docker-compose.yml` (merge into your existing one),
   `.env.example` (merge), and `scripts/run-api.sh` /
   `scripts/forge-deploy.sh` (adapt to your existing deploy script) into your
   app's repo.
2. Run `scripts/new-app.sh` first (before copying, in a scratch clone of this
   template) to set your module path, app name, and cookie names — or do the
   equivalent renames by hand in `backend/` after copying.
3. Point your frontend's API base at the Go service the same way
   `frontend/app/composables/useApi.ts` does here: `credentials: 'include'`
   on every request, and echo the CSRF cookie (`<slug>_csrf` by default) back
   as an `X-CSRF-Token` header on non-GET requests. Port that composable
   directly — it has no dependency on the rest of `frontend/`.
4. Add the auth flow to your existing router: a login page that POSTs
   `/api/auth/magic-link`, and a callback route that POSTs
   `/api/auth/verify` with the `?token=` query param and then reloads the
   current user (`GET /api/me`). `frontend/app/pages/login.vue` and
   `frontend/app/pages/auth/callback.vue` are directly portable if your app
   also happens to be Nuxt.
5. In `docker-compose.yml`, keep the `postgres` and `mailpit` services and
   the `backend` service's build/env block; drop the `frontend` service if
   your monorepo already defines its own.
6. Same-domain cookies require your frontend and the Go API to share a
   parent domain (or be proxied under one origin) in production — see
   `docs/DEPLOY.md`'s Nginx section for the reverse-proxy shape this assumes.

## What you get

Everything under `backend/internal/`: `auth` (magic link + sessions + CSRF),
`users` (allowlist), `mailer` (Resend/SMTP/SES), `config`, `db` (pgx pool +
migrations + sqlc), and `httpapi` (chi router, rate limiting, security
headers). None of it imports anything from `frontend/`, so it composes with
any frontend stack, not just Nuxt — the auth contract is just cookies and
two JSON endpoints.
