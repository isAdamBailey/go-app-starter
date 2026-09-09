#!/usr/bin/env bash
# Rewrites this template repo in place for a new app: Go module path, app
# name/title, cookie names, PM2 process name, and default DB/user names.
#
# Usage:
#   ./scripts/new-app.sh github.com/you/new-app "New App" new_app
#
# Run this once, right after cloning the template for a new project, then
# commit the result. Safe to re-run, but it's meant to run exactly once.
set -euo pipefail

if [[ $# -ne 3 ]]; then
  echo "Usage: $0 <new-go-module-path> <\"Display Name\"> <db_and_env_slug>" >&2
  echo "Example: $0 github.com/you/new-app \"New App\" new_app" >&2
  exit 1
fi

NEW_MODULE="$1"        # e.g. github.com/you/new-app/backend
NEW_NAME="$2"           # e.g. "New App"
NEW_SLUG="$3"           # e.g. new_app (used for DB name, cookie prefix, PM2 process)

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

OLD_MODULE="github.com/isAdamBailey/go-app-starter/backend"

echo "Rewriting Go module path: $OLD_MODULE -> $NEW_MODULE"
grep -rl "$OLD_MODULE" backend | while read -r f; do
  sed -i.bak "s#$OLD_MODULE#$NEW_MODULE#g" "$f" && rm "$f.bak"
done

echo "Renaming cookies: app_session/app_csrf -> ${NEW_SLUG}_session/${NEW_SLUG}_csrf"
sed -i.bak "s#\"app_session\"#\"${NEW_SLUG}_session\"#; s#\"app_csrf\"#\"${NEW_SLUG}_csrf\"#" backend/internal/auth/cookies.go && rm backend/internal/auth/cookies.go.bak
sed -i.bak "s#'app_csrf'#'${NEW_SLUG}_csrf'#" frontend/app/composables/useApi.ts && rm frontend/app/composables/useApi.ts.bak

echo "Setting app name: $NEW_NAME"
sed -i.bak "s#appName           = \"the app\"#appName           = \"$NEW_NAME\"#" backend/internal/mailer/mailer.go && rm backend/internal/mailer/mailer.go.bak
sed -i.bak "s#title: 'App'#title: '$NEW_NAME'#" frontend/nuxt.config.ts && rm frontend/nuxt.config.ts.bak
sed -i.bak "s#>\s*App\s*<#>$NEW_NAME<#g" frontend/app/pages/login.vue frontend/app/pages/auth/callback.vue frontend/app/pages/index.vue && \
  rm -f frontend/app/pages/login.vue.bak frontend/app/pages/auth/callback.vue.bak frontend/app/pages/index.vue.bak

echo "Renaming DB defaults: app -> $NEW_SLUG"
sed -i.bak "s#:-app}#:-${NEW_SLUG}}#g; s#app:app@postgres#${NEW_SLUG}:${NEW_SLUG}@postgres#g; s#login@app.local#login@${NEW_SLUG}.local#g" docker-compose.yml .env.example && \
  rm docker-compose.yml.bak .env.example.bak

echo "Renaming PM2 process: app-web -> ${NEW_SLUG}-web"
sed -i.bak "s#app-web#${NEW_SLUG}-web#g" frontend/ecosystem.config.cjs scripts/forge-deploy.sh && \
  rm frontend/ecosystem.config.cjs.bak scripts/forge-deploy.sh.bak

echo "Regenerating go.sum for the new module path..."
(cd backend && go mod tidy)

cat <<EOF

Done. Remaining manual steps:
  1. Review the diff (git diff) — a few TODO comments still need your input
     (package description, favicon, meta tags, PM2 name collisions).
  2. Update package.json's "name" field in frontend/ if you care about it.
  3. Update docs/DEPLOY.md's example domain/paths for your real deployment.
  4. Set a real COOKIE_SIGNING_SECRET in .env (openssl rand -base64 32).
  5. Rename this repo itself (git remote / GitHub repo name) if desired.
EOF
