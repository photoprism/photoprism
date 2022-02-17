#!/usr/bin/env bash

# abort if the user is not root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run install-tensorflow.sh as root" 1>&2
  exit 1
fi

if [[ -z $1 ]] && [[ -z $2 ]]; then
    echo "Usage: install-tensorflow.sh [amd64|arm64|arm] [cpu|avx|avx2]" 1>&2
    exit 1
else
    echo "install-tensorflow: downloading library..."
    if [[ $1 == "amd64" ]]; then
        TARGETARCH="linux"
        TARGETVARIANT="-${2:-"cpu"}"
        URL="https://dl.photoprism.app/tensorflow/$TARGETARCH/libtensorflow-${TARGETARCH}${TARGETVARIANT}-1.15.2.tar.gz"
    elif [[ $1 == "arm64" ]]; then
        URL="https://dl.photoprism.app/tensorflow/$1/libtensorflow-$1-1.15.2.tar.gz"
    elif [[ $1 == "arm" ]]; then
        URL="https://dl.photoprism.app/tensorflow/$1/libtensorflow-$1-1.15.2.tar.gz"
    else
        echo "install-tensorflow: unsupported architecture" 1>&2
        exit 1
    fi
    echo "$URL"
    curl -fsSL "$URL" | \
    tar --overwrite -C "/usr" -xz && \
    ldconfig
    echo "install-tensorflow: done"
fi
