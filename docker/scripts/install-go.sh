#!/usr/bin/env bash

set -e

GOLANG_VERSION=1.17.3

if [[ -z $1 ]]; then
    echo "architecture required: amd64, arm64, or arm" 1>&2
    exit 1
else
    set -eux;
    if [[ $1 == "amd64" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-$1.tar.gz"
        CHECKSUM="550f9845451c0c94be679faf116291e7807a8d78b43149f9506c1b15eb89008c *go.tgz"
    elif [[ $1 == "arm64" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-$1.tar.gz"
        CHECKSUM="06f505c8d27203f78706ad04e47050b49092f1b06dc9ac4fbee4f0e4d015c8d4 *go.tgz"
    elif [[ $1 == "arm" ]]; then
        URL="https://golang.org/dl/go${GOLANG_VERSION}.linux-armv6l.tar.gz"
        CHECKSUM="aa0d5516c8bd61654990916274d27491cfa229d322475502b247a8dc885adec5 *go.tgz"
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
