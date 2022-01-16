#!/usr/bin/env bash

set -e

GOLANG_VERSION=1.17.6

if [[ -z $1 ]]; then
    echo "architecture required: amd64, arm64, or arm" 1>&2
    exit 1
else
    set -eux;
    if [[ $1 == "amd64" ]]; then
        URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"
        CHECKSUM="231654bbf2dab3d86c1619ce799e77b03d96f9b50770297c8f4dff8836fc8ca2 *go.tgz"
    elif [[ $1 == "arm64" ]]; then
        URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-arm64.tar.gz"
        CHECKSUM="82c1a033cce9bc1b47073fd6285233133040f0378439f3c4659fe77cc534622a *go.tgz"
    elif [[ $1 == "arm" ]]; then
        URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
        CHECKSUM="9ac723e6b41cb7c3651099a09332a8a778b69aa63a5e6baaa47caf0d818e2d6d *go.tgz"
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
