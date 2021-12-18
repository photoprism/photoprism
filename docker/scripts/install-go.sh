#!/usr/bin/env bash

set -e

GOLANG_VERSION=1.17.5

if [[ -z $1 ]]; then
    echo "architecture required: amd64, arm64, or arm" 1>&2
    exit 1
else
    set -eux;
    if [[ $1 == "amd64" ]]; then
        URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"
        CHECKSUM="bd78114b0d441b029c8fe0341f4910370925a4d270a6a590668840675b0c653e *go.tgz"
    elif [[ $1 == "arm64" ]]; then
        URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-arm64.tar.gz"
        CHECKSUM="6f95ce3da40d9ce1355e48f31f4eb6508382415ca4d7413b1e7a3314e6430e7e *go.tgz"
    elif [[ $1 == "arm" ]]; then
        URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
        CHECKSUM="aa1fb6c53b4fe72f159333362a10aca37ae938bde8adc9c6eaf2a8e87d1e47de *go.tgz"
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
