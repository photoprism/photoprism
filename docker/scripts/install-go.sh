#!/usr/bin/env bash

set -e

GOLANG_VERSION=1.17.4

if [[ -z $1 ]]; then
    echo "architecture required: amd64, arm64, or arm" 1>&2
    exit 1
else
    set -eux;
    if [[ $1 == "amd64" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-$1.tar.gz"
        CHECKSUM="adab2483f644e2f8a10ae93122f0018cef525ca48d0b8764dae87cb5f4fd4206 *go.tgz"
    elif [[ $1 == "arm64" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-$1.tar.gz"
        CHECKSUM="617a46bd083e59877bb5680998571b3ddd4f6dcdaf9f8bf65ad4edc8f3eafb13 *go.tgz"
    elif [[ $1 == "arm" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
        CHECKSUM="f34d25f818007345b716b316908115ba462f3f0dbd4343073020b544b4ae2320 *go.tgz"
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
