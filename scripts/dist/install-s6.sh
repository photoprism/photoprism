#!/usr/bin/env bash

# This downloads and installs the s6-overlay binaries and noarch files from GitHub:
# - https://github.com/just-containers/s6-overlay
#
# s6 is a suite of utilities for UNIX that provides process supervision, e.g. for use with Docker:
# - https://github.com/skarnet/s6
# - https://skarnet.org/software/s6/
# - https://ahmet.im/blog/minimal-init-process-for-containers/
# - https://labex.io/tutorials/docker-how-to-gracefully-shut-down-a-long-running-docker-container-417742

# Show usage information if first argument is --help.
if [[ ${1} == "--help" ]]; then
  echo "Usage: ${0##*/} [version] [dir]" 1>&2
  exit 0
fi

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "${0##*/} [version] [dir] must be run as root" 1>&2
  exit 1
fi

# You can provide a custom installation directory as the first argument.
S6_OVERLAY_DESTDIR=$(realpath "${2:-/}")

# Determine target architecture.
if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

# shellcheck source=/dev/null
. /etc/os-release

S6_OVERLAY_ARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

case $S6_OVERLAY_ARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    S6_OVERLAY_ARCH=x86_64
    ;;

  arm64 | ARM64 | aarch64)
    S6_OVERLAY_ARCH=aarch64
    ;;

  arm | ARM | aarch | armv7l | armhf)
    S6_OVERLAY_ARCH=armhf
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$S6_OVERLAY_ARCH\"" 1>&2
    exit 1
    ;;
esac

set -eu

S6_OVERLAY_LATEST=$(curl --silent "https://api.github.com/repos/just-containers/s6-overlay/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
S6_OVERLAY_VERSION=${1:-$S6_OVERLAY_LATEST}
ARCHIVE_NOARCH="s6-overlay-noarch.tar.xz"
ARCHIVE_BINARY="s6-overlay-${S6_OVERLAY_ARCH}.tar.xz"
S6_NOARCH_URL="https://github.com/just-containers/s6-overlay/releases/download/${S6_OVERLAY_VERSION}/${ARCHIVE_NOARCH}"
S6_BINARY_URL="https://github.com/just-containers/s6-overlay/releases/download/${S6_OVERLAY_VERSION}/${ARCHIVE_BINARY}"

echo "Installing S6 Overlay for ${S6_OVERLAY_ARCH^^}..."

echo "--------------------------------------------------------------------------------"
echo "VERSION: ${S6_OVERLAY_VERSION}"
echo "LATEST : ${S6_OVERLAY_LATEST}"
echo "NOARCH : ${ARCHIVE_NOARCH}"
echo "BINARY : ${ARCHIVE_BINARY}"
echo "DESTDIR: ${S6_OVERLAY_DESTDIR}"
echo "--------------------------------------------------------------------------------"

# Create the destination directory if it does not already exist.
mkdir -p "${S6_OVERLAY_DESTDIR}"

# Stage downloads in a temporary directory that is removed on exit.
S6_TMPDIR=$(mktemp -d)
trap 'rm -rf "${S6_TMPDIR}"' EXIT

# verify_sha256 checks a file against the SHA-256 published next to the release
# asset (<url>.sha256), aborting on mismatch. It soft-fails when no checksum is
# published so pinned older tags without a manifest still install.
verify_sha256() {
  local url="$1" file="$2" sumfile="$3" expected actual
  if ! curl -fsSL "${url}.sha256" -o "${sumfile}" 2>/dev/null; then
    echo "Warning: no published checksum at ${url}.sha256; skipping verification." 1>&2
    return 0
  fi
  expected=$(awk '{print $1}' "${sumfile}")
  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "${file}" | awk '{print $1}')
  else
    actual=$(shasum -a 256 "${file}" | awk '{print $1}')
  fi
  if [[ ${expected} != "${actual}" ]]; then
    echo "Error: SHA-256 mismatch for $(basename "${file}")." 1>&2
    echo "  expected: ${expected}" 1>&2
    echo "  actual:   ${actual}" 1>&2
    exit 1
  fi
  echo "Checksum OK ($(basename "${file}"): ${actual})."
}

# download_and_extract fetches a release tarball, verifies it, then extracts it.
download_and_extract() {
  local url="$1" name="$2" archive="${S6_TMPDIR}/${2}"
  echo "Downloading \"${url}\"..."
  curl -fsSL "${url}" -o "${archive}"
  verify_sha256 "${url}" "${archive}" "${archive}.sha256"
  echo "Extracting \"${name}\" to \"${S6_OVERLAY_DESTDIR}\"..."
  tar -C "${S6_OVERLAY_DESTDIR}" -Jxp -f "${archive}"
}

# Download, verify, and install the s6-overlay release from GitHub.
download_and_extract "$S6_NOARCH_URL" "$ARCHIVE_NOARCH"
download_and_extract "$S6_BINARY_URL" "$ARCHIVE_BINARY"

echo "Done."
