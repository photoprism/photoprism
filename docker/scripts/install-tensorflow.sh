#!/usr/bin/env bash

if [[ -z $1 ]] && [[ -z $2 ]]; then
    echo "Please define architecture and variant (cpu/avx2/...)" 1>&2
    exit 1
else
    echo "Downloading tensorflow library..."
    if [[ $1 == "amd64" ]]; then
        TARGETARCH="linux"
        TARGETVARIANT="-${2:-"cpu"}"
        URL="https://dl.photoprism.org/tensorflow/$TARGETARCH/libtensorflow-${TARGETARCH}${TARGETVARIANT}-1.15.2.tar.gz"
    elif [[ $1 == "arm64" ]]; then
        URL="https://dl.photoprism.org/tensorflow/$1/libtensorflow-$1-1.15.2.tar.gz"
    elif [[ $1 == "arm" ]]; then
        URL="https://github.com/Qengineering/TensorFlow-Raspberry-Pi/blob/master/libtensorflow_1_15_2.tar.gz?raw=true"
    else
        echo "cpu architecture not supported by now" 1>&2
        exit 1
    fi
    echo "$URL"
    curl -L "$URL" | \
    tar -C "/usr" -xz && \
    ldconfig
    echo "Done"
fi
