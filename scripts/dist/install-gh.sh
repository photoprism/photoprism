#!/usr/bin/env bash

# Installs the latest GitHub CLI on Debian/Ubuntu from cli.github.com.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-gh.sh)

set -Eeuo pipefail
IFS=$'\n\t'

PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"

if [[ ! -r /etc/os-release ]]; then
  echo "ERROR: /etc/os-release not found." >&2
  exit 1
fi

# shellcheck source=/dev/null
. /etc/os-release

if [[ "${ID:-}" != "debian" && "${ID:-}" != "ubuntu" ]] && [[ " ${ID_LIKE:-} " != *" debian "* ]]; then
  echo "ERROR: This installer supports Debian/Ubuntu only." >&2
  exit 1
fi

if ! command -v apt-get >/dev/null 2>&1; then
  echo "ERROR: apt-get not found." >&2
  exit 1
fi

SUDO=""
if [[ "$(id -u)" -ne 0 ]]; then
  if command -v sudo >/dev/null 2>&1; then
    SUDO="sudo"
  else
    echo "ERROR: root or sudo access is required." >&2
    exit 1
  fi
fi

export DEBIAN_FRONTEND="noninteractive"

if ! command -v wget >/dev/null 2>&1; then
  ${SUDO} apt-get update
  ${SUDO} apt-get install -y wget
fi

${SUDO} mkdir -p -m 755 /etc/apt/keyrings
keyring_tmp="$(mktemp)"
trap 'rm -f "${keyring_tmp}"' EXIT
wget -nv -O "${keyring_tmp}" https://cli.github.com/packages/githubcli-archive-keyring.gpg
${SUDO} tee /etc/apt/keyrings/githubcli-archive-keyring.gpg < "${keyring_tmp}" >/dev/null
${SUDO} chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg

${SUDO} mkdir -p -m 755 /etc/apt/sources.list.d
printf 'deb [arch=%s signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main\n' "$(dpkg --print-architecture)" \
  | ${SUDO} tee /etc/apt/sources.list.d/github-cli.list >/dev/null

${SUDO} apt-get update
${SUDO} apt-get install -y gh

if command -v gh >/dev/null 2>&1; then
  gh --version | head -n 1
fi
