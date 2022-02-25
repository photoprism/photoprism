#!/usr/bin/env bash

if [[ -z $1 ]] || [[ -z $2 ]]; then
  echo "Usage: ${0##*/} [debug|race|static|prod] [filename]" 1>&2
  exit 1
fi

set -e

BUILD_OS=$(uname -s)
BUILD_ARCH=$("$(dirname "$0")/dist/arch.sh")
BUILD_DATE=$(date -u +%y%m%d)
BUILD_VERSION=$(git describe --always)
BUILD_TAG=${BUILD_DATE}-${BUILD_VERSION}
BUILD_ID=${BUILD_TAG}-${BUILD_OS}-${BUILD_ARCH^^}

echo "Building PhotoPrism ${BUILD_ID} ($1)..."

if [[ $1 == "debug" ]]; then
  go build -ldflags "-X main.version=${BUILD_ID}-DEBUG" -o $2 cmd/photoprism/photoprism.go
  du -h $2
elif [[ $1 == "race" ]]; then
  go build -race -ldflags "-X main.version=${BUILD_ID}-DEBUG" -o $2 cmd/photoprism/photoprism.go
  du -h $2
elif [[ $1 == "static" ]]; then
  go build -a -v -ldflags "-linkmode external -extldflags \"-static -L /usr/lib -ltensorflow\" -s -w -X main.version=${BUILD_ID}" -o $2 cmd/photoprism/photoprism.go
  du -h $2
else
  go build -ldflags "-extldflags \"-Wl,-rpath -Wl,\$ORIGIN/../lib\" -s -w -X main.version=${BUILD_ID}" -o $2 cmd/photoprism/photoprism.go
  du -h $2
fi

echo "Done."