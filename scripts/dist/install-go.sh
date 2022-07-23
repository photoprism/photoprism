#!/bin/bash

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

GOLANG_VERSION=1.18.4
DESTDIR=$(realpath "${1:-/usr/local}")

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

echo "Installing Go in \"$DESTDIR\"..."

set -e

if command -v arch &> /dev/null
then
    SYSTEM_ARCH=$(arch)
elif [[ $PHOTOPRISM_ARCH ]]; then
    SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
    SYSTEM_ARCH=$("$(dirname "$0")/arch.sh")
fi

DESTARCH=${2:-$SYSTEM_ARCH}

mkdir -p "$DESTDIR"

set -eux;

if [[ $DESTARCH == "amd64" ]]; then
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"
    CHECKSUM="c9b099b68d93f5c5c8a8844a89f8db07eaa58270e3a1e01804f17f4cf8df02f5 *go.tgz"
elif [[ $DESTARCH == "arm64" ]]; then
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-arm64.tar.gz"
    CHECKSUM="35014d92b50d97da41dade965df7ebeb9a715da600206aa59ce1b2d05527421f *go.tgz"
elif [[ $DESTARCH == "arm" ]]; then
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
    CHECKSUM="7dfeab572e49638b0f3d9901457f0622c27b73301c2b99db9f5e9568ff40460c *go.tgz"
else
    echo "Unsupported Machine Architecture: $DESTARCH" 1>&2
    exit 1
fi

echo "Downloading Go from \"$URL\". Please wait."

wget -O go.tgz $URL
echo "$CHECKSUM" | sha256sum -c -
rm -rf /usr/local/go
tar -C /usr/local -xzf go.tgz
rm go.tgz

/usr/local/go/bin/go version

echo "Done."