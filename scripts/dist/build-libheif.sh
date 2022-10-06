#!/usr/bin/env bash

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

# Build "heif-convert", "heif-enc", "heif-info", and "heif-thumbnailer" from source.
CURRENT_DIR=$(pwd)
apt-get update
apt-get -qq install git autoconf automake cmake libtool libjpeg8 libjpeg8-dev libde265-dev
cd "/tmp" || exit
rm -rf "/tmp/libheif"
git clone https://github.com/strukturag/libheif.git
cd libheif || exit
./autogen.sh
./configure
make

# Install "heif-convert", "heif-enc", "heif-info", and "heif-thumbnailer" in "/usr/local".
make install-exec
cd "$CURRENT_DIR" || exit
rm -rf "/tmp/libheif"

# Create a tar archive to distribute the binaries on demand.
if [[ $1 ]]; then
    echo "creating $1..."
    (cd /usr/local && tar -czf "$1" lib/libheif.* bin/heif-convert bin/heif-enc bin/heif-info bin/heif-thumbnailer)
fi

