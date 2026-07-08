#!/usr/bin/env bash

# Installs the hadolint Dockerfile linter on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-hadolint.sh)
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-hadolint.sh) -- --version v2.14.0
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-hadolint.sh) -- /opt/photoprism

set -euo pipefail

REPO="hadolint/hadolint"
TOOL="hadolint"

# Show usage information if first argument is --help.
if [[ ${1:-} == "--help" ]]; then
  echo "Usage: ${0##*/} [--version <tag>] [destdir]" 1>&2
  echo "" 1>&2
  echo "Environment:" 1>&2
  echo "  PHOTOPRISM_ARCH=amd64|arm64" 1>&2
  exit 0
fi

VERSION="latest"

while [[ $# -gt 0 ]]; do
  case $1 in
    --version)
      VERSION=${2:-}
      if [[ -z $VERSION ]]; then
        echo "Error: --version requires a tag (e.g. v2.14.0)." 1>&2
        exit 1
      fi
      shift 2
      ;;
    --help)
      echo "Usage: ${0##*/} [--version <tag>] [destdir]" 1>&2
      exit 0
      ;;
    --)
      shift
      break
      ;;
    -*)
      echo "Error: Unknown option: $1" 1>&2
      exit 1
      ;;
    *)
      break
      ;;
  esac
done

# You can provide a custom installation directory as the first argument.
DESTDIR=$(realpath "${1:-/usr/local}")

# Determine target architecture.
if [[ -n ${PHOTOPRISM_ARCH:-} ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi
DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    ASSET="hadolint-linux-x86_64"
    ;;
  arm64 | ARM64 | aarch64)
    ASSET="hadolint-linux-arm64"
    ;;
  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

# Abort if not executed as root when installing into a system directory.
if [[ $(id -u) != "0" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Error: Run ${0##*/} as root to install in a system directory!" 1>&2
  exit 1
fi

if [[ $VERSION == "latest" ]]; then
  RELEASE_URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"
  CHECKSUM_URL="${RELEASE_URL}.sha256"
else
  RELEASE_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"
  CHECKSUM_URL="${RELEASE_URL}.sha256"
fi

DESTBIN="${DESTDIR}/bin/${TOOL}"

echo "--------------------------------------------------------------------------------"
echo "VERSION  : ${VERSION}"
echo "ASSET    : ${ASSET}"
echo "DOWNLOAD : ${RELEASE_URL}"
echo "DESTBIN  : ${DESTBIN}"
echo "--------------------------------------------------------------------------------"

echo "Downloading ${TOOL} to \"${DESTBIN}\"..."
mkdir -p "${DESTDIR}/bin"
tmp_bin=$(mktemp)
trap 'rm -f "${tmp_bin}" "${tmp_bin}.sha256"' EXIT

curl --fail --silent --show-error --location "${RELEASE_URL}" -o "${tmp_bin}"

# Verify the SHA-256 checksum when the upstream releases publish one.
if curl --fail --silent --show-error --location "${CHECKSUM_URL}" -o "${tmp_bin}.sha256" 2>/dev/null; then
  expected=$(awk '{print $1}' "${tmp_bin}.sha256")
  actual=$(sha256sum "${tmp_bin}" | awk '{print $1}')
  if [[ "${expected}" != "${actual}" ]]; then
    echo "Error: SHA-256 mismatch for ${TOOL} (${ASSET})." 1>&2
    echo "  expected: ${expected}" 1>&2
    echo "  actual:   ${actual}" 1>&2
    exit 1
  fi
  echo "Checksum OK (${actual})."
else
  echo "Warning: no published checksum found at ${CHECKSUM_URL}; skipping verification." 1>&2
fi

install -m 755 "${tmp_bin}" "${DESTBIN}"

if command -v "${TOOL}" >/dev/null 2>&1; then
  "${TOOL}" --version | head -n 1
fi

echo "Done."
