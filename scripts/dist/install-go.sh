#!/bin/bash

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts"

GOLANG_VERSION=1.18.1
DESTDIR=$(realpath "${1:-/usr/local}")

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

echo "Installing Go in \"$DESTDIR\"..."

set -e

SYSTEM_ARCH=$("$(dirname "$0")/arch.sh")
DESTARCH=${2:-$SYSTEM_ARCH}

mkdir -p "$DESTDIR"

set -eux;

if [[ $DESTARCH == "amd64" ]]; then
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"
    CHECKSUM="b3b815f47ababac13810fc6021eb73d65478e0b2db4b09d348eefad9581a2334 *go.tgz"
elif [[ $DESTARCH == "arm64" ]]; then
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-arm64.tar.gz"
    CHECKSUM="56a91851c97fb4697077abbca38860f735c32b38993ff79b088dac46e4735633 *go.tgz"
elif [[ $DESTARCH == "arm" ]]; then
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
    CHECKSUM="9edc01c8e7db64e9ceeffc8258359e027812886ceca3444e83c4eb96ddb068ee *go.tgz"
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