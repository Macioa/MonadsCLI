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
}

build_and_install ubuntu "$ROOT_DIR/test/docker/Dockerfile.ubuntu"
build_and_install debian "$ROOT_DIR/test/docker/Dockerfile.debian"
build_and_install alpine "$ROOT_DIR/test/docker/Dockerfile.alpine"
build_and_install fedora "$ROOT_DIR/test/docker/Dockerfile.fedora"
build_and_install ubuntu-pwsh "$ROOT_DIR/test/docker/Dockerfile.ubuntu-pwsh" "linux/amd64"
