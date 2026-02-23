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

  # Verify install succeeded (CLI tools present) and Go tests pass
  # Skip agent/copilot on Alpine: their native binaries are glibc-only (fcntl64 symbol not found on musl)
  docker run --rm "${platform_args[@]}" "monadscli-test:${name}-installed" \
    /bin/sh -c '
      set -e
      export PATH="$HOME/.local/bin:$PATH"
      [ -f /etc/alpine-release ] || agent --version
      gemini --version
      claude -v
      [ -f /etc/alpine-release ] || copilot --version
      qodo_output="$(qodo --version 2>&1 || true)"
      [ -n "$qodo_output" ] || (echo "qodo --version produced no output" && exit 1)
      cd /app && go test ./...
    '
}

run_one() {
  case "$1" in
    ubuntu)       run_tests ubuntu ;;
    debian)       run_tests debian ;;
    alpine)       run_tests alpine ;;
    fedora)       run_tests fedora ;;
    ubuntu-pwsh)  run_tests ubuntu-pwsh "linux/amd64" ;;
    *) echo "Unknown distro: $1" >&2; exit 1 ;;
  esac
}

if [[ -n "${1:-}" ]]; then
  run_one "$1"
else
  run_one ubuntu
  run_one debian
  run_one alpine
  run_one fedora
  run_one ubuntu-pwsh
fi
