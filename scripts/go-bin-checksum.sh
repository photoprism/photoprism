#!/usr/bin/env bash

# checksums for current go version 1.16.4 (linux)
if [[ -z $1 ]]; then
    echo "Please define architecture and version" 1>&2
    exit 1
else
    if [[ $1 == "amd64" ]]; then
        echo "7154e88f5a8047aad4b80ebace58a059e36e7e2e4eb3b383127a28c711b4ff59 *go.tgz"
    elif [[ $1 == "arm64" ]]; then
        echo "8b18eb05ddda2652d69ab1b1dd1f40dd731799f43c6a58b512ad01ae5b5bba21 *go.tgz"
    elif [[ $1 == "arm" ]]; then
        echo "a53391a800ddec749ee90d38992babb27b95cfb864027350c737b9aa8e069494 *go.tgz"
    else
        echo "cpu architecture not supported by now" 1>&2
        exit 1
    fi
fi
