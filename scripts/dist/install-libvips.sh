#!/usr/bin/env bash

# Installs libvips from a Jammy backport PPA.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-libvips.sh)

set -Eeuo pipefail
IFS=$'\n\t'

PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"

if [[ ! -r /etc/os-release ]]; then
  echo "ERROR: /etc/os-release not found." >&2
  exit 1
fi

# shellcheck source=/dev/null
. /etc/os-release

if [[ "${ID:-}" != "ubuntu" ]]; then
  echo "ERROR: This installer currently supports Ubuntu only." >&2
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

PPA_USER="${LIBVIPS_PPA_USER:-0k53d-karl-f830m}"
PPA_NAME="${LIBVIPS_PPA_NAME:-vips}"
PPA_FINGERPRINT="${LIBVIPS_PPA_FINGERPRINT:-573634A0CBF3F3DCECCF1EB7212ED20BE4FE4BA6}"
PPA_BASE_URL="https://ppa.launchpadcontent.net/${PPA_USER}/${PPA_NAME}/ubuntu"
PPA_CODENAME="${VERSION_CODENAME:-jammy}"
KEYRING_PATH="/etc/apt/keyrings/libvips-archive-keyring.gpg"
SOURCES_PATH="/etc/apt/sources.list.d/libvips-ppa.list"

export DEBIAN_FRONTEND="noninteractive"

if ! command -v curl >/dev/null 2>&1 || ! command -v gpg >/dev/null 2>&1; then
  ${SUDO} apt-get update
  ${SUDO} apt-get install -y curl gnupg ca-certificates
fi

${SUDO} mkdir -p -m 755 /etc/apt/keyrings

key_tmp="$(mktemp)"
keyring_tmp="$(mktemp)"
trap 'rm -f "${key_tmp}" "${keyring_tmp}"' EXIT

curl -fsSL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x${PPA_FINGERPRINT}" -o "${key_tmp}"

found_fingerprint="$(gpg --show-keys --with-colons "${key_tmp}" | awk -F: '/^fpr:/ { print $10; exit }')"
if [[ "${found_fingerprint}" != "${PPA_FINGERPRINT}" ]]; then
  echo "ERROR: Unexpected signing key fingerprint '${found_fingerprint}'." >&2
  exit 1
fi

gpg --dearmor < "${key_tmp}" > "${keyring_tmp}"
${SUDO} install -m 644 -o root -g root "${keyring_tmp}" "${KEYRING_PATH}"

${SUDO} mkdir -p -m 755 /etc/apt/sources.list.d
printf 'deb [arch=%s signed-by=%s] %s %s main\n' \
  "$(dpkg --print-architecture)" "${KEYRING_PATH}" "${PPA_BASE_URL}" "${PPA_CODENAME}" \
  | ${SUDO} tee "${SOURCES_PATH}" >/dev/null

${SUDO} apt-get update
${SUDO} apt-get install -y libvips-dev

if command -v vips >/dev/null 2>&1; then
  vips --version
fi
