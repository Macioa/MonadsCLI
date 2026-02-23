#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

export DOCKER_BUILDKIT=0

build_and_install() {
  local name="$1"
  local dockerfile="$2"
  local platform="${3:-}"
  local platform_args=()

  if [[ -n "$platform" ]]; then
    platform_args=(--platform "$platform")
  fi

  docker build "${platform_args[@]}" -f "$dockerfile" -t "monadscli-test:${name}" "$ROOT_DIR"

  local container="monadscli-install-${name}"
  docker run --name "$container" "${platform_args[@]}" "monadscli-test:${name}"
  docker commit "$container" "monadscli-test:${name}-installed" >/dev/null
  docker rm "$container" >/dev/null

  # Free disk on CI: drop pre-install image (run-tests only needs -installed) and dangling layers
  docker rmi "monadscli-test:${name}" 2>/dev/null || true
  docker image prune -f
}

run_one() {
  case "$1" in
    ubuntu)       build_and_install ubuntu "$ROOT_DIR/test/docker/Dockerfile.ubuntu" ;;
    debian)       build_and_install debian "$ROOT_DIR/test/docker/Dockerfile.debian" ;;
    alpine)       build_and_install alpine "$ROOT_DIR/test/docker/Dockerfile.alpine" ;;
    fedora)       build_and_install fedora "$ROOT_DIR/test/docker/Dockerfile.fedora" ;;
    ubuntu-pwsh)  build_and_install ubuntu-pwsh "$ROOT_DIR/test/docker/Dockerfile.ubuntu-pwsh" "linux/amd64" ;;
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
