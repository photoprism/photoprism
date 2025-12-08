#!/usr/bin/env bash

# Installs the Codex CLI coding agent on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-codex.sh)

set -Eeuo pipefail

echo "Installing Codex CLI..."

# Ensure npm exists
if ! command -v npm >/dev/null 2>&1; then
  echo "ERROR: npm not found. Please install Node.js (npm) and re-run." >&2
  exit 1
fi

# Create CODEX_HOME if set (and not '/')
if [ -n "${CODEX_HOME:-}" ]; then
  if [ "${CODEX_HOME}" = "/" ]; then
    echo "ERROR: refusing to use CODEX_HOME='/'" >&2
    exit 2
  fi
  install -d -m 700 -- "${CODEX_HOME}"
fi

# Choose sudo only if available and not already root
SUDO=""
if command -v sudo >/dev/null 2>&1 && [ "$(id -u)" -ne 0 ]; then
  SUDO="sudo"
fi

# Some npm versions don’t support --location=global; detect and adapt
if npm help install 2>/dev/null | grep -q -- '--location'; then
  NPM_GLOBAL_OPTS=(install -g --location=global --no-fund --no-audit)
else
  NPM_GLOBAL_OPTS=(install -g --no-fund --no-audit)
fi

# Install / update Codex CLI
$SUDO npm "${NPM_GLOBAL_OPTS[@]}" "@openai/codex@latest"

# Show result
if command -v codex >/dev/null 2>&1; then
  echo "Codex installed at: $(command -v codex)"
  codex --version || true
fi

echo "Done."
