#!/usr/bin/env bash

# This installs the heif-convert, heif-enc, heif-info, and heif-thumbnailer binaries on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-libheif.sh)

set -e

# Show usage information if first argument is --help.
if [[ ${1} == "--help" ]]; then
  echo "Usage: ${0##*/} [destdir] [version]" 1>&2
  exit 0
fi

# You can provide a custom installation directory as the first argument.
DESTDIR=$(realpath "${1:-/usr/local}")

# In addition, you can specify a custom version to be installed as the second argument.
LIBHEIF_VERSION=${2:-v1.18.1}

# Determine the system architecture.
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
    DESTARCH=arm
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

mkdir -p "$DESTDIR"

# Map codenames to find and use a compatible version.
case $VERSION_CODENAME in
  vera | virginia)
    VERSION_CODENAME=jammy
    ;;
esac

echo "Installing libheif..."

ARCHIVE="libheif-${VERSION_CODENAME}-${DESTARCH}-${LIBHEIF_VERSION}.tar.gz"
URL="https://dl.photoprism.app/dist/libheif/${ARCHIVE}"

echo "------------------------------------------------"
echo "VERSION: $LIBHEIF_VERSION"
echo "ARCHIVE: $ARCHIVE"
echo "DESTDIR: $DESTDIR"
echo "------------------------------------------------"

echo "Extracting \"$URL\" to \"$DESTDIR\"."
wget --inet4-only -c "$URL" -O - | tar --overwrite --mode=755 -xz -C "$DESTDIR"

if [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Running \"ldconfig\"."
  ldconfig
else
  echo "Running \"ldconfig -n $DESTDIR/lib\"."
  ldconfig -n "$DESTDIR/lib"
fi

echo "Done."
