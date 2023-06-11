#!/usr/bin/env bash

set -e

DESTDIR=$(realpath "${1:-/usr/local}")
LIBHEIF_VERSION=${2:-v1.13.0}
SYSTEM_ARCH=$("$(dirname "$0")/arch.sh")
DESTARCH=${DESTARCH:-$SYSTEM_ARCH}

. /etc/os-release

# Abort if not executed as root.
if [[ $(id -u) != "0" ]] && [[ $DESTDIR == "/usr" || $DESTDIR == "/usr/local" ]]; then
  echo "Error: Run ${0##*/} as root to install in a system directory!" 1>&2
  exit 1
fi

. /etc/os-release

mkdir -p "$DESTDIR"

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
