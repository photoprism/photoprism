#!/usr/bin/env bash

GOLANG_VERSION=1.16.6

if [[ -z $1 ]]; then
    echo "Please define architecture and version" 1>&2
    exit 1
else
    set -eux;
    if [[ $1 == "amd64" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-$1.tar.gz"
        CHECKSUM="be333ef18b3016e9d7cb7b1ff1fdb0cac800ca0be4cf2290fe613b3d069dfe0d *go.tgz"
    elif [[ $1 == "arm64" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-$1.tar.gz"
        CHECKSUM="9e38047463da6daecab9017cd0599f33f84991e68263752cfab49253bbc98c30 *go.tgz"
    elif [[ $1 == "arm" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
        CHECKSUM="b1ca342e81897da3f25da4e75ae29b267db1674fe7222d9bfc4c666bcf6fce69 *go.tgz"
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
