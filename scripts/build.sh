#!/usr/bin/env bash

set -e

BUILD_DATE=$(date -u +%y%m%d)
BUILD_VERSION=$(git describe --always)

if [[ -z $1 ]] || [[ -z $2 ]]; then
  echo "Usage: build.sh [debug|race|static|prod] [filename]" 1>&2
  exit 1
fi

if [[ $OS == "Windows_NT" ]]; then
  BUILD_OS=win32
  if [[ $PROCESSOR_ARCHITEW6432 == "AMD64" ]]; then
    BUILD_ARCH=amd64
  else
    if [[ $PROCESSOR_ARCHITECTURE == "AMD64" ]]; then
      BUILD_ARCH=amd64
    fi
    if [[ $PROCESSOR_ARCHITECTURE == "x86" ]]; then
      BUILD_ARCH=ia32
    fi
  fi
else
  BUILD_OS=$(uname -s)
  BUILD_ARCH=$(uname -m)
fi

BUILD_ID=${BUILD_DATE}-${BUILD_VERSION}-${BUILD_OS}-${BUILD_ARCH}

echo "Building $1 binary..."
echo "Version: PhotoPrism CE ${BUILD_ID}"

if [[ $1 == "debug" ]]; then
  go build -ldflags "-X main.version=${BUILD_ID}-DEBUG" -o $2 cmd/photoprism/photoprism.go
  du -h $2
  echo "Done."
elif [[ $1 == "race" ]]; then
  go build -race -ldflags "-X main.version=${BUILD_ID}-DEBUG" -o $2 cmd/photoprism/photoprism.go
  du -h $2
elif [[ $1 == "static" ]]; then
  go build -a -v -ldflags "-linkmode external -extldflags \"-static -L /usr/lib -ltensorflow\" -s -w -X main.version=${BUILD_ID}" -o $2 cmd/photoprism/photoprism.go
  du -h $2
else
  go build -ldflags "-s -w -X main.version=${BUILD_ID}" -o $2 cmd/photoprism/photoprism.go
  du -h $2
fi

echo "Done."