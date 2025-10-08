#!/usr/bin/env bash

# Installs the qdrant binary, a vector search engine, on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-qdrant.sh)

set -e

# Show usage information if first argument is --help.
if [[ ${1} == "--help" ]]; then
  echo "Usage: ${0##*/} [destdir] [version]" 1>&2
  exit 0
fi

# You can provide a custom installation directory as the first argument.
DESTDIR=$(realpath "${1:-/usr/local}")

# Determine target architecture.
if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    DESTARCH=x86_64
    ;;

  arm64 | ARM64 | aarch64)
    DESTARCH=aarch64
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

# shellcheck source=/dev/null
. /etc/os-release

# Abort if not executed as root.
if [[ $(id -u) != "0" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Error: Run ${0##*/} as root to install in a system directory!" 1>&2
  exit 1
fi

echo "Installing Qdrant for ${DESTARCH^^}..."

# Alternatively, users can specify a custom version to install as the second argument.
GITHUB_LATEST=$(curl --silent "https://api.github.com/repos/qdrant/qdrant/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
VERSION=${2:-$GITHUB_LATEST}
ARCHIVE="qdrant-${DESTARCH}-unknown-linux-musl.tar.gz"
GITHUB_URL="https://github.com/qdrant/qdrant/releases/download/${VERSION}/${ARCHIVE}"

echo "--------------------------------------------------------------------------------"
echo "VERSION : ${VERSION}"
echo "LATEST  : ${GITHUB_LATEST}"
echo "DOWNLOAD: ${GITHUB_URL}"
echo "DESTDIR : ${DESTDIR}"
echo "--------------------------------------------------------------------------------"

# Adjust the installation path because the archive does not contain a bin directory.
DESTDIR="${DESTDIR}/bin"

echo "Extracting the qdrant binary in \"${ARCHIVE}\" to \"${DESTDIR}\"..."
mkdir -p "${DESTDIR}"
curl -fsSL "${GITHUB_URL}" | tar --overwrite --mode=755 -xz -C "${DESTDIR}" --wildcards --no-anchored "qdrant"

echo "Done."
