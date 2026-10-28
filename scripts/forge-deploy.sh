#!/usr/bin/env bash
# Builds a Forge zero-downtime release in place. The Forge deploy script runs
# this from the new release directory, between $CREATE_RELEASE() and
# $ACTIVATE_RELEASE(), then restarts the API daemon itself:
#
#   $CREATE_RELEASE()
#   cd $FORGE_RELEASE_DIRECTORY
#   bash scripts/forge-deploy.sh
#   $ACTIVATE_RELEASE()
#   sudo supervisorctl restart daemon-<id>:*
#
# See docs/DEPLOY.md.
set -euo pipefail

ROOT="$PWD"

if [[ ! -f "$ROOT/scripts/forge-deploy.sh" ]]; then
  echo "Run this from the repo root of the release (got $ROOT)" >&2
  exit 1
fi

export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"

# Forge keeps .env at the site root, two levels above releases/<id>.
ENV_FILE=""
for candidate in "$ROOT/../../.env" "$ROOT/frontend/.env" "$ROOT/.env"; do
  if [[ -f "$candidate" ]]; then
    ENV_FILE="$(cd "$(dirname "$candidate")" && pwd)/$(basename "$candidate")"
    break
  fi
done

if [[ -z "$ENV_FILE" ]]; then
  echo "No .env found (checked $ROOT/../../.env, $ROOT/frontend/.env and $ROOT/.env)" >&2
  exit 1
fi

echo "Using env from $ENV_FILE"
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

# Builds from every app on the server share one lock so two deploys don't
# run Go and Node builds at once and exhaust memory.
exec 9>/tmp/node-build.lock
flock 9

export GOFLAGS="-p=2"

cd "$ROOT/backend"
go build -o server ./cmd/server
# Migrate before activation so a bad migration fails the deploy and the
# current release keeps serving.
go run ./cmd/migrate

cd "$ROOT/frontend"
npm ci
npm run generate
