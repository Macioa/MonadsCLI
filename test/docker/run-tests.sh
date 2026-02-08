#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

run_tests() {
  local name="$1"
  local platform="${2:-}"
  local platform_args=()

  if [[ -n "$platform" ]]; then
    platform_args=(--platform "$platform")
  fi

  docker run --rm "${platform_args[@]}" "monadscli-test:${name}-installed" \
    /bin/sh -c "cd /app && go test ./..."
}

run_tests ubuntu
run_tests debian
run_tests alpine
run_tests fedora
run_tests ubuntu-pwsh "linux/amd64"
