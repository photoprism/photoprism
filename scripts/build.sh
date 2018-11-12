#!/usr/bin/env bash

BUILD_DATE=`date -u +%y%m%d`
VERSION=`git describe --always`

if [[ -z $1 ]] || [[ -z $2 ]]; then
    echo "Please provide build mode and output file name" 1>&2
    exit 1
fi

if [[ $OS == "Windows_NT" ]]; then
    OPERATING_SYSTEM=win32
    if [[ $PROCESSOR_ARCHITEW6432 == "AMD64" ]]; then
        PROCESSOR=amd64
    else
        if [[ $PROCESSOR_ARCHITECTURE == "AMD64" ]]; then
            PROCESSOR=amd64
        fi
        if [[ $PROCESSOR_ARCHITECTURE == "x86" ]]; then
            PROCESSOR=ia32
        fi
    fi
else
    OPERATING_SYSTEM=`uname -s`
    PROCESSOR=`uname -p`
fi

if [[ $1 == "debug" ]]; then
    echo "Building development binary..."
	go build -ldflags "-X main.version=${BUILD_DATE}-${VERSION}-${OPERATING_SYSTEM}-${PROCESSOR}-DEBUG" -o $2 cmd/photoprism/photoprism.go
	du -h $2
	echo "Done."
else
    echo "Building production binary..."
	go build -ldflags "-s -w -X main.version=${BUILD_DATE}-${VERSION}-${OPERATING_SYSTEM}-${PROCESSOR}" -o $2 cmd/photoprism/photoprism.go
	du -h $2
	echo "Done."
fi