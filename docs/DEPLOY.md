# Deploying

Production runs on the shared Go platform: a **Laravel Forge** server with
PostgreSQL 18 on the box, one Forge site per app, and the Go API as a Forge
daemon. Nuxt is generated to static files at deploy time, so there is **no
Node process** in production.

Placeholders used below:

| Placeholder | Example |
| --- | --- |
| `<domain>` | `yourapp.example.com` |
| `<slug>` | `your_app` (the slug passed to `scripts/new-app.sh`) |
| `<port>` | `8103` (from the port registry below) |
| `<id>` | the Forge daemon id, e.g. `1234567` |

## Architecture

```
https://<domain>
        │
        ▼
     nginx (Forge-managed)
        ├─ /api/*, /healthz → 127.0.0.1:<port>   (Go daemon)
        └─ /*               → current/frontend/.output/public (static files)

Postgres → 127.0.0.1:5432 on the same server, one role + database per app
Email    → SES over SMTP (port 2587)
```

Cookie auth requires **one origin**: nginx serves the generated SPA and
proxies `/api` to Go. `NUXT_PUBLIC_API_BASE` stays empty in production so the
frontend calls same-origin `/api`.

---

## 1. Port registry

Every app's Go daemon gets its own port from the **81xx registry**
(face-value#18): face-value `8101`, massa `8102`, the next app `8103`, and so
on. `PORT` is always set explicitly in `.env`, never defaulted — the code
default (`8080`) is for local dev only.

Check the port is free before claiming it:

```sh
sudo ss -ltnp | grep ':8103' || echo free
```

Add the app to the registry in face-value#18 once it's assigned.

---

## 2. Postgres role and database

Each app gets its own role (with a password, no superuser) and its own
database that no other role can connect to:

```sh
sudo -u postgres createuser --pwprompt --no-superuser --no-createdb --no-createrole <slug>
sudo -u postgres createdb -O <slug> <slug>
sudo -u postgres psql -c "REVOKE CONNECT ON DATABASE <slug> FROM PUBLIC;"
sudo -u postgres psql -c "GRANT CONNECT ON DATABASE <slug> TO <slug>;"
```

Connect over TCP to `127.0.0.1`, not the Unix socket:

```
DATABASE_URL=postgres://<slug>:PASSWORD@127.0.0.1:5432/<slug>?sslmode=disable
```

`pgcrypto` is a trusted extension, so the app role creates it from the
migrations itself — no superuser step needed.

---

## 3. Create the Forge site

1. **Sites → New Site** → `<domain>`.
2. Project type: **Nuxt**, with the build type set to **Static**.
3. **Web directory:** `/frontend/.output/public` (Forge prefixes `current/`
   for zero-downtime sites; the nginx `root` in step 7 is what counts).
4. Enable **zero-downtime deployments** (`releases/<id>` + a `current`
   symlink).
5. Connect **GitHub** → your repo, branch `main`. Leave **push to deploy** off
   until the first deploy succeeds.
6. **SSL** → Let's Encrypt. Forge keeps the custom nginx blocks from step 7
   when it rewrites the config for the certificate.

---

## 4. Environment

Forge → site → **Environment**. Forge writes this to
`/home/forge/<domain>/.env`, at the **site root** — shared by every release.
`scripts/forge-deploy.sh` reads it from `../../.env` (relative to
`releases/<id>`) and `scripts/run-api.sh` from `../.env` (relative to
`current/`). Both fail loudly if it's missing.

| Variable | Local (`.env.example`) | Production |
| --- | --- | --- |
| `PORT` | `8080` | `<port>` from the registry, e.g. `8103` |
| `DATABASE_URL` | `postgres://…@postgres:5432/…` | `postgres://<slug>:…@127.0.0.1:5432/<slug>?sslmode=disable` |
| `APP_BASE_URL` | `http://localhost:3000` | `https://<domain>` |
| `COOKIE_SIGNING_SECRET` | dev placeholder | `openssl rand -base64 32` |
| `COOKIE_SECURE` | `false` | `true` |
| `EMAIL_PROVIDER` | `smtp` | `ses` (or `resend`) |
| `SES_REGION` | empty | your SES region, e.g. `us-west-2` |
| `SMTP_HOST` | `mailpit` | omit (derived from `SES_REGION`) |
| `SMTP_PORT` | `1025` | **`2587`** — this server can't open 587 or 465; SES answers on 2587 |
| `SMTP_USERNAME` / `SMTP_PASSWORD` | empty | SES SMTP credentials |
| `MAGIC_LINK_FROM_EMAIL` | `login@app.local` | your verified sender |
| `ALLOWED_EMAILS` | your email | same — must match what you type at login |
| `NUXT_PUBLIC_API_BASE` | `http://localhost:8080` | **empty** (same-origin `/api`; baked in at generate time) |

Leave the code's `SMTP_PORT` default (`587`) alone; `2587` is a property of
this server, so it lives in `.env`.

Changing a backend variable only needs a daemon restart
(`sudo supervisorctl restart daemon-<id>:*`). Changing `NUXT_PUBLIC_*` needs a
redeploy, because it's baked into the generated files.

If the app stores encrypted data, its encryption key in `.env` must move with
the database, or the data is unreadable. Compare keys across servers by
fingerprint instead of pasting them:

```sh
grep '^KEY=' .env | cut -d= -f2- | sha256sum | cut -c1-12
```

---

## 5. Deploy script

Forge → site → **Deployments → Deploy script**:

```bash
$CREATE_RELEASE()

cd $FORGE_RELEASE_DIRECTORY

bash scripts/forge-deploy.sh

$ACTIVATE_RELEASE()

sudo supervisorctl restart daemon-<id>:*
```

`scripts/forge-deploy.sh` runs inside the new release, before it's
activated. It:

1. loads `.env` (site root, falling back to `frontend/.env` and `.env`) and
   fails if none exists;
2. takes `flock` on `/tmp/node-build.lock` so deploys of different apps on
   this server don't build at the same time, and sets `GOFLAGS=-p=2`;
3. builds `backend/server`;
4. runs the migrations — a bad migration fails the deploy and the old release
   keeps serving;
5. runs `npm ci` and `npm run generate`.

The `:*` suffix on the restart matters: Forge daemons are supervisor groups,
and `daemon-<id>` without it isn't found.

When moving an existing app onto this script, turn **push to deploy off** on
the old site before merging the rewrite, or the old site deploys the new
script against the old shape.

---

## 6. Go API daemon

**Server → Daemons → New Daemon**:

| Field | Value |
| --- | --- |
| Command | `bash scripts/run-api.sh` |
| Directory | `/home/forge/<domain>/current` |
| User | `forge` |

Note the daemon id for the deploy script's restart line.

---

## 7. nginx

Forge → site → **Nginx**. Replace the generated `root` and `location /`
blocks with:

```nginx
root /home/forge/<domain>/current/frontend/.output/public;
index index.html;

location /api/ {
    proxy_pass http://127.0.0.1:<port>;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 60s;
}

location = /healthz {
    proxy_pass http://127.0.0.1:<port>;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
}

location /_nuxt/ {
    expires 1y;
    add_header Cache-Control "public, max-age=31536000, immutable";
    try_files $uri =404;
}

location / {
    try_files $uri $uri/ /200.html;
}
```

`200.html` is the SPA fallback `npm run generate` writes; client-side routes
that aren't prerendered fall through to it.

`backend/internal/httpapi/clientip.go` only trusts `X-Forwarded-For` when the
connecting peer is loopback (this nginx), so per-IP rate limiting sees real
client IPs.

---

## 8. DNS

Point the domain at the server. DigitalOcean's nameservers lag panel edits
by about 5 minutes — wait that out before debugging a stale answer.

---

## 9. First deploy

1. **Deploy Now** in Forge.
2. The log should show the `.env` path, the Go build, `migrations applied`,
   `npm ci` and `nuxt generate`, then the daemon restart.
3. Verify:

   ```sh
   curl -s https://<domain>/healthz
   ls /home/forge/<domain>/current/frontend/.output/public/{index,200}.html
   sudo supervisorctl status | grep daemon-<id>
   ```

4. Request a magic link and complete login.
5. Turn **push to deploy** on.

---

## 10. Backups

The nightly `pg-backup.sh` Forge scheduled job dumps every non-template
database to S3, so a new app is covered without any change. Check the new
database shows up in S3 the next morning.

The job must run as **root**. It once ran as `forge` for two nights and
failed silently with `Permission denied` (face-value#18, massa#15). After the
first scheduled run, read `~/.forge/scheduled-<id>.log` — a successful hand
run doesn't prove the scheduled one works.

---

## Troubleshooting

| Symptom | Fix |
| --- | --- |
| 502 on `/api` | The daemon is down or on another port. `sudo supervisorctl status`; run `bash scripts/run-api.sh` from `current/` to see the error; check `PORT` in `.env` matches the nginx `proxy_pass`. |
| Deploy: `daemon-<id>: ERROR (no such process)` | Add the `:*` suffix: `sudo supervisorctl restart daemon-<id>:*`. |
| Daemon log: `bind: address already in use` | Another app holds the port. Check `sudo ss -ltnp`, pick a free port from the registry, update `PORT` and the nginx blocks. |
| Magic link lands on the login screen | nginx redirected to `/auth/callback/`. The auth middleware strips trailing slashes (`normalizePath` in `frontend/app/middleware/auth.global.ts`); make sure that change is deployed. Also check `APP_BASE_URL` is the HTTPS domain and `COOKIE_SECURE=true`. |
| Magic link never arrives | Email not in `ALLOWED_EMAILS`, or SMTP is blocked: `SMTP_PORT` must be `2587` on this server. Check the daemon log. |
| Deploy: `No .env found` | Set the environment in Forge so `/home/forge/<domain>/.env` exists at the site root. |
| Deploy fails on `go build` | Install Go to `/usr/local/go`, matching `backend/go.mod`. |
| Deploy fails on migrate | Fix the migration; the previous release is still live because activation never ran. |

**Diagnose magic links:**

```sh
sudo -u postgres psql -d <slug> -c \
  "SELECT user_email, created_at FROM magic_link_tokens ORDER BY created_at DESC LIMIT 5;"
sudo -u postgres psql -d <slug> -c "SELECT email FROM allowed_users;"
sudo supervisorctl tail -100 daemon-<id>:daemon-<id>_00
```

No token rows → address not allowed or rate-limited (5/hour, plus 5/min/IP —
see `httpapi.magicLinkRateLimit`). Token rows but no mail → check SMTP
credentials, `SMTP_PORT=2587`, and the sender address.
