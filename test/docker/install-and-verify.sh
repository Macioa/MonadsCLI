#!/usr/bin/env bash
set -euo pipefail

export PATH="$HOME/.local/bin:$PATH"
export PIP_BREAK_SYSTEM_PACKAGES=1

monadscli install

gemini --version
agent --version
claude -v
copilot --version

if command -v aider >/dev/null 2>&1; then
  aider --version
else
  python -m aider --version
fi

qodo_output="$(qodo --version 2>&1 || true)"
if [[ -z "$qodo_output" ]]; then
  echo "qodo --version produced no output"
  exit 1
fi
echo "$qodo_output"
