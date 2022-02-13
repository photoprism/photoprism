#!/usr/bin/env bash

set -e

GOLANG_VERSION=1.17.7

if [[ -z $1 ]]; then
    echo "architecture required: amd64, arm64, or arm" 1>&2
    exit 1
else
    set -eux;
    if [[ $1 == "amd64" ]]; then
        URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"
        CHECKSUM="02b111284bedbfa35a7e5b74a06082d18632eff824fd144312f6063943d49259 *go.tgz"
    elif [[ $1 == "arm64" ]]; then
        URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-arm64.tar.gz"
        CHECKSUM="a5aa1ed17d45ee1d58b4a4099b12f8942acbd1dd09b2e9a6abb1c4898043c5f5 *go.tgz"
    elif [[ $1 == "arm" ]]; then
        URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
        CHECKSUM="874774f078b182fa21ffcb3878467eb5cb7e78bbffa6343ea5f0fbe47983433b *go.tgz"
    else
        echo "unsupported architecture" 1>&2
        exit 1
    fi
    wget -O go.tgz $URL
    echo $CHECKSUM | sha256sum -c -
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go.tgz
    rm go.tgz
    /usr/local/go/bin/go version
fi
