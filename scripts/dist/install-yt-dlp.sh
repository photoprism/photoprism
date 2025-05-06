#!/usr/bin/env bash

# Installs the yt-dlp binary, a vector search engine, on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-yt-dlp.sh)

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

  arm | ARM | aarch | armv7l | armhf)
    DESTARCH=armv7l
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

echo "Installing yt-dlp for ${DESTARCH^^}..."

# Alternatively, users can specify a custom version to install as the second argument.
GITHUB_LATEST=$(curl --silent "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
VERSION=${2:-$GITHUB_LATEST}
BINARY="yt-dlp_linux"

if [[ $DESTARCH != "x86_64" ]]; then
  BINARY="${BINARY}_${DESTARCH}"
fi

GITHUB_URL="https://github.com/yt-dlp/yt-dlp/releases/download/${VERSION}/${BINARY}"
DESTBIN="${DESTDIR}/bin/yt-dlp"

echo "------------------------------------------------"
echo "VERSION : ${VERSION}"
echo "LATEST  : ${GITHUB_LATEST}"
echo "DOWNLOAD: ${GITHUB_URL}"
echo "DESTDIR : ${DESTDIR}"
echo "DESTBIN : ${DESTBIN}"
echo "------------------------------------------------"

echo "Downloading the yt-dlp binary to \"${DESTBIN}\"..."
mkdir -p "${DESTDIR}"
curl -fsSL "${GITHUB_URL}" -o "${DESTBIN}"

echo "Changing permissions of \"${DESTBIN}\" to 755..."
chmod 755 "${DESTBIN}"

echo "Done."
