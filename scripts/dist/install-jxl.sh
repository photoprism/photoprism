#!/usr/bin/env bash

# This installs JPEG XL on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-jxl.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

LIB_VERSION=${2:-v0.8.1}
SYSTEM_ARCH=$("$(dirname "$0")/arch.sh")
DESTARCH=${DESTARCH:-$SYSTEM_ARCH}

set -e

. /etc/os-release

ARCHIVE="jxl-debs-${DESTARCH}-ubuntu-22.04-${LIB_VERSION}.tar.gz"
URL="https://github.com/libjxl/libjxl/releases/download/${LIB_VERSION}/${ARCHIVE}"
TMPDIR="/tmp/jpegxl"

echo "------------------------------------------------"
echo "VERSION: $LIB_VERSION"
echo "ARCHIVE: $ARCHIVE"
echo "------------------------------------------------"

echo "Installing JPEG XL for ${DESTARCH^^}..."

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    if [[ $VERSION_CODENAME == "jammy" ]]; then
      apt-get update
      apt-get install -f libtcmalloc-minimal4 libhwy-dev libhwy0
      rm -rf /tmp/jpegxl
      mkdir -p "$TMPDIR"
      echo "Extracting \"$URL\" to \"$TMPDIR\"."
      wget --inet4-only -c "$URL" -O - | tar --overwrite --mode=755 -xz -C "$TMPDIR"
      (cd "$TMPDIR" && dpkg -i jxl_0.8.1_amd64.deb libjxl_0.8.1_amd64.deb libjxl-dev_0.8.1_amd64.deb)
      apt --fix-broken install
      rm -rf /tmp/jpegxl
    else
      echo "install-jxl: target distribution currently unsupported"
    fi
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$BUILD_ARCH\"" 1>&2
    exit 0
    ;;
esac

echo "Done."
