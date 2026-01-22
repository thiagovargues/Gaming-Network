#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ ! -f "$ROOT_DIR/backend/.env.local" ]]; then
  echo "Missing backend/.env.local. Create it first." >&2
  exit 1
fi

backend_pid=""

cleanup() {
  if [[ -n "$backend_pid" ]] && kill -0 "$backend_pid" 2>/dev/null; then
    kill "$backend_pid" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

(
  set -a
  # shellcheck disable=SC1091
  source "$ROOT_DIR/backend/.env.local"
  set +a
  cd "$ROOT_DIR/backend"
  mkdir -p "$ROOT_DIR/logs"
  go run ./cmd/api 2>&1 | tee "$ROOT_DIR/logs/backend.log"
) &
backend_pid=$!

( cd "$ROOT_DIR/frontend" && PORT=3000 npm run dev | tee "$ROOT_DIR/logs/frontend.log" )
