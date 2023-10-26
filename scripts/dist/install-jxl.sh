#!/usr/bin/env bash

# This installs JPEG XL on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-jxl.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

# Determine the system architecture.
if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

LIB_VERSION=${2:-v0.8.1}
SYSTEM_ARCH=$("$(dirname "$0")/arch.sh")

set -e

. /etc/os-release

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    if [[ $VERSION_CODENAME == "jammy" ]]; then
      ARCHIVE="jxl-debs-${DESTARCH}-ubuntu-22.04-${LIB_VERSION}.tar.gz"
      URL="https://github.com/libjxl/libjxl/releases/download/${LIB_VERSION}/${ARCHIVE}"
      TMPDIR="/tmp/jpegxl"

      echo "------------------------------------------------"
      echo "VERSION: $LIB_VERSION"
      echo "ARCHIVE: $ARCHIVE"
      echo "------------------------------------------------"

      echo "Installing JPEG XL for ${DESTARCH^^}..."

      apt-get update
      apt-get install -f libtcmalloc-minimal4 libhwy-dev libhwy0
      rm -rf /tmp/jpegxl
      mkdir -p "$TMPDIR"
      echo "Extracting \"$URL\" to \"$TMPDIR\"."
      wget --inet4-only -c "$URL" -O - | tar --overwrite --mode=755 -xz -C "$TMPDIR"
      (cd "$TMPDIR" && dpkg -i jxl_0.8.1_amd64.deb libjxl_0.8.1_amd64.deb libjxl-dev_0.8.1_amd64.deb)
      apt --fix-broken install
      rm -rf /tmp/jpegxl
    elif [[ $VERSION_CODENAME == "lunar" || $VERSION_CODENAME == "mantic" ]]; then
      echo "Installing JPEG XL distribution packages for amd64 (Intel 64-bit)"
      apt-get -qq install libjxl-dev libjxl-tools
    else
      echo "JPEG XL is currently unsupported."
    fi
    ;;

  arm64 | ARM64 | aarch64)
    if [[ $VERSION_CODENAME == "lunar" || $VERSION_CODENAME == "mantic" ]]; then
      echo "Installing JPEG XL distribution packages for arm64 (ARM 64-bit)"
      apt-get -qq install libjxl-dev libjxl-tools
    else
      echo "JPEG XL is currently unsupported."
    fi
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$DESTARCH\"" 1>&2
    exit 0
    ;;
esac

echo "Done."
