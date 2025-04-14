#!/usr/bin/env bash

# Installs the nats-server binary, a cloud-native messaging system, on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-nats.sh)

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
    DESTARCH=amd64
    ;;

  arm64 | ARM64 | aarch64)
    DESTARCH=arm64
    ;;

  arm | ARM | aarch | armv7l | armhf)
    DESTARCH=arm7
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 1
    ;;
esac

. /etc/os-release

# Abort if not executed as root.
if [[ $(id -u) != "0" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Error: Run ${0##*/} as root to install in a system directory!" 1>&2
  exit 1
fi

echo "Installing NATS for ${DESTARCH^^}..."

# Alternatively, users can specify a custom version to install as the second argument.
GITHUB_LATEST=$(curl --silent "https://api.github.com/repos/nats-io/nats-server/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
VERSION=${2:-$GITHUB_LATEST}
ARCHIVE="nats-server-${VERSION}-linux-${DESTARCH}.tar.gz"
GITHUB_URL="https://github.com/nats-io/nats-server/releases/download/${VERSION}/${ARCHIVE}"

echo "------------------------------------------------"
echo "VERSION : ${VERSION}"
echo "LATEST  : ${GITHUB_LATEST}"
echo "DOWNLOAD: ${GITHUB_URL}"
echo "DESTDIR : ${DESTDIR}"
echo "------------------------------------------------"

# Adjust the installation path because the archive does not contain a bin directory.
DESTDIR="${DESTDIR}/bin"

echo "Extracting the nats-server binary in \"${ARCHIVE}\" to \"${DESTDIR}\"..."
mkdir -p "${DESTDIR}"
curl -fsSL "${GITHUB_URL}" | tar --overwrite --mode=755 -xz -C "${DESTDIR}" --strip-components=1 --wildcards --no-anchored "nats-server"

echo "Done."
