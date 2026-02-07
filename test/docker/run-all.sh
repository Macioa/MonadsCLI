#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

export DOCKER_BUILDKIT=0

build_and_run() {
  local name="$1"
  local dockerfile="$2"
  local platform="${3:-}"
  local platform_args=()

  if [[ -n "$platform" ]]; then
    platform_args=(--platform "$platform")
  fi

  docker build "${platform_args[@]}" -f "$dockerfile" -t "monadscli-test:${name}" "$ROOT_DIR"
  docker run --rm "${platform_args[@]}" "monadscli-test:${name}"
}

build_and_run ubuntu "$ROOT_DIR/test/docker/Dockerfile.ubuntu"
build_and_run debian "$ROOT_DIR/test/docker/Dockerfile.debian"
build_and_run alpine "$ROOT_DIR/test/docker/Dockerfile.alpine"
build_and_run fedora "$ROOT_DIR/test/docker/Dockerfile.fedora"
build_and_run ubuntu-pwsh "$ROOT_DIR/test/docker/Dockerfile.ubuntu-pwsh" "linux/amd64"
