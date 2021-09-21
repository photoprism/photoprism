#!/usr/bin/env bash

set -e

GOLANG_VERSION=1.17.1

if [[ -z $1 ]]; then
    echo "architecture required: amd64, arm64, or arm" 1>&2
    exit 1
else
    set -eux;
    if [[ $1 == "amd64" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-$1.tar.gz"
        CHECKSUM="dab7d9c34361dc21ec237d584590d72500652e7c909bf082758fb63064fca0ef *go.tgz"
    elif [[ $1 == "arm64" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-$1.tar.gz"
        CHECKSUM="53b29236fa03ed862670a5e5e2ab2439a2dc288fe61544aa392062104ac0128c *go.tgz"
    elif [[ $1 == "arm" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
        CHECKSUM="ed3e4dbc9b80353f6482c441d65b51808290e94ff1d15d56da5f4a7be7353758 *go.tgz"
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
