#!/usr/bin/env bash

GOLANG_VERSION=1.17.7

DESTDIR=$(realpath "${1:-/usr/local}")

echo "Installing Go in \"$DESTDIR\"..."

# abort if the user is not root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run install-go.sh as root" 1>&2
  exit 1
fi

set -e

mkdir -p "$DESTDIR"

SYSTEM_ARCH=$("$(dirname "$0")/arch.sh")
INSTALL_ARCH=${2:-$SYSTEM_ARCH}

set -eux;

if [[ $INSTALL_ARCH == "amd64" ]]; then
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"
    CHECKSUM="02b111284bedbfa35a7e5b74a06082d18632eff824fd144312f6063943d49259 *go.tgz"
elif [[ $INSTALL_ARCH == "arm64" ]]; then
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-arm64.tar.gz"
    CHECKSUM="a5aa1ed17d45ee1d58b4a4099b12f8942acbd1dd09b2e9a6abb1c4898043c5f5 *go.tgz"
elif [[ $INSTALL_ARCH == "arm" ]]; then
    URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
    CHECKSUM="874774f078b182fa21ffcb3878467eb5cb7e78bbffa6343ea5f0fbe47983433b *go.tgz"
else
    echo "Unsupported Machine Architecture: $INSTALL_ARCH" 1>&2
    exit 1
fi

echo "Downloading Go from \"$URL\". Please wait."

wget -O go.tgz $URL
echo $CHECKSUM | sha256sum -c -
rm -rf /usr/local/go
tar -C /usr/local -xzf go.tgz
rm go.tgz

/usr/local/go/bin/go version

echo "Done."