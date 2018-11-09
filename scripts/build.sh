#!/usr/bin/env bash

VERSION=`date -u +0.%Y%m%d.%H%M%S`
BRANCH=`git rev-parse --abbrev-ref HEAD`

if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Please provide build mode and output file name" 1>&2
    exit 1
fi

if [ $1 == "debug" ]; then
    echo "Building development binary..."
	go build -ldflags "-X main.version=${VERSION}-${BRANCH}" -o $2 cmd/photoprism/photoprism.go
	echo "Done."
else
    echo "Building production binary..."
	go build -ldflags "-s -w -X main.version=${VERSION}" -o $2 cmd/photoprism/photoprism.go
	echo "Done."
fi