#!/bin/bash

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts"

GOLANG_VERSION=1.18.2
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
    CHECKSUM="e54bec97a1a5d230fc2f9ad0880fcbabb5888f30ed9666eca4a91c5a32e86cbc *go.tgz"
elif [[ $DESTARCH == "arm64" ]]; then
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-arm64.tar.gz"
    CHECKSUM="fc4ad28d0501eaa9c9d6190de3888c9d44d8b5fb02183ce4ae93713f67b8a35b *go.tgz"
elif [[ $DESTARCH == "arm" ]]; then
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
    CHECKSUM="570dc8df875b274981eaeabe228d0774985de42e533ffc8c7ff0c9a55174f697 *go.tgz"
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