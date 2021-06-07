#!/usr/bin/env bash

GOLANG_VERSION=1.16.4

if [[ -z $1 ]]; then
    echo "Please define architecture and version" 1>&2
    exit 1
else
    set -eux;
    if [[ $1 == "amd64" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-$1.tar.gz"
        CHECKSUM="7154e88f5a8047aad4b80ebace58a059e36e7e2e4eb3b383127a28c711b4ff59 *go.tgz"
    elif [[ $1 == "arm64" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-$1.tar.gz"
        CHECKSUM="8b18eb05ddda2652d69ab1b1dd1f40dd731799f43c6a58b512ad01ae5b5bba21 *go.tgz"
    elif [[ $1 == "arm" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
        CHECKSUM="a53391a800ddec749ee90d38992babb27b95cfb864027350c737b9aa8e069494 *go.tgz"
    else
        echo "cpu architecture not supported by now" 1>&2
        exit 1
    fi
    wget -O go.tgz $URL
    echo $CHECKSUM | sha256sum -c -
    tar -C /usr/local -xzf go.tgz
    rm go.tgz
    go version
    mkdir -p "/go/src" "/go/bin" && chmod -R 777 "/go"
fi
