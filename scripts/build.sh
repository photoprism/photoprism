#!/usr/bin/env bash

VERSION=`date -u +0.%Y%m%d.%H%M%S`
BRANCH=`git rev-parse --abbrev-ref HEAD`

if [ ${BRANCH} == "master" ]; then
    echo "Building production binary..."
	go build -ldflags "-s -w -X main.version=${VERSION}" cmd/photoprism/photoprism.go
	echo "Done."
else
    echo "Building development binary..."
	go build -ldflags "-X main.version=${VERSION}-${BRANCH}" cmd/photoprism/photoprism.go
	echo "Done."
fi