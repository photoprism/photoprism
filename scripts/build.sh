#!/usr/bin/env bash

PHOTOPRISM_DATE=`date -u +%y%m%d`
PHOTOPRISM_VERSION=`git describe --always`

if [[ -z $1 ]] || [[ -z $2 ]]; then
    echo "Please provide build mode and output file name" 1>&2
    exit 1
fi

if [[ $OS == "Windows_NT" ]]; then
    PHOTOPRISM_OS=win32
    if [[ $PROCESSOR_ARCHITEW6432 == "AMD64" ]]; then
        PHOTOPRISM_ARCH=amd64
    else
        if [[ $PROCESSOR_ARCHITECTURE == "AMD64" ]]; then
            PHOTOPRISM_ARCH=amd64
        fi
        if [[ $PROCESSOR_ARCHITECTURE == "x86" ]]; then
            PHOTOPRISM_ARCH=ia32
        fi
    fi
else
    PHOTOPRISM_OS=`uname -s`
    PHOTOPRISM_ARCH=`uname -p`
fi

if [[ $1 == "debug" ]]; then
  echo "Building development binary..."
	go build -ldflags "-X main.version=${PHOTOPRISM_DATE}-${PHOTOPRISM_VERSION}-${PHOTOPRISM_OS}-${PHOTOPRISM_ARCH}-DEBUG" -o $2 cmd/photoprism/photoprism.go
	du -h $2
	echo "Done."
elif [[ $1 == "static" ]]; then
  echo "Building static production binary..."
	go build -a -v -ldflags "-linkmode external -extldflags \"-static -L /usr/lib -ltensorflow\" -s -w -X main.version=${PHOTOPRISM_DATE}-${PHOTOPRISM_VERSION}-${PHOTOPRISM_OS}-${PHOTOPRISM_ARCH}" -o $2 cmd/photoprism/photoprism.go
	du -h $2
	echo "Done."
else
  echo "Building production binary..."
	go build -ldflags "-s -w -X main.version=${PHOTOPRISM_DATE}-${PHOTOPRISM_VERSION}-${PHOTOPRISM_OS}-${PHOTOPRISM_ARCH}" -o $2 cmd/photoprism/photoprism.go
	du -h $2
	echo "Done."
fi
