#!/bin/bash

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts"

GOLANG_VERSION=1.18.3
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
    CHECKSUM="956f8507b302ab0bb747613695cdae10af99bbd39a90cae522b7c0302cc27245 *go.tgz"
elif [[ $DESTARCH == "arm64" ]]; then
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-arm64.tar.gz"
    CHECKSUM="beacbe1441bee4d7978b900136d1d6a71d150f0a9bb77e9d50c822065623a35a *go.tgz"
elif [[ $DESTARCH == "arm" ]]; then
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
    CHECKSUM="b8f0b5db24114388d5dcba7ca0698510ea05228b0402fcbeb0881f74ae9cb83b *go.tgz"
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