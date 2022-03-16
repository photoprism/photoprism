#!/usr/bin/env bash

if [[ -z $1 ]] || [[ -z $2 ]]; then
  echo "Usage: ${0##*/} [debug|race|static|prod] [filename]" 1>&2
  exit 1
fi

set -e

BUILD_OS=$(/bin/uname -s)
BUILD_ARCH=$("$(/usr/bin/dirname "$0")/dist/arch.sh")
BUILD_DATE=$(/bin/date -u +%y%m%d)
BUILD_VERSION=$(git describe --always)
BUILD_TAG=${BUILD_DATE}-${BUILD_VERSION}
BUILD_ID=${BUILD_TAG}-${BUILD_OS}-${BUILD_ARCH^^}
BUILD_NAME=${2:-photoprism}
GO_BIN=${GO_BIN:-/usr/local/go/bin/go}
GO_VER=$($GO_BIN version)

echo "Building PhotoPrism ${BUILD_ID} ($1)..."

if [[ $1 == "debug" ]]; then
  BUILD_CMD=("$GO_BIN" build -ldflags "-X main.version=${BUILD_ID}-DEBUG" -o "${BUILD_NAME}" cmd/photoprism/photoprism.go)
elif [[ $1 == "race" ]]; then
  BUILD_CMD=("$GO_BIN" build -race -ldflags "-X main.version=${BUILD_ID}-DEBUG" -o "${BUILD_NAME}" cmd/photoprism/photoprism.go)
elif [[ $1 == "static" ]]; then
  BUILD_CMD=("$GO_BIN" build -a -v -ldflags "-linkmode external -extldflags \"-static -L /usr/lib -ltensorflow\" -s -w -X main.version=${BUILD_ID}" -o "${BUILD_NAME}" cmd/photoprism/photoprism.go)
else
  BUILD_CMD=("$GO_BIN" build -ldflags "-extldflags \"-Wl,-rpath -Wl,\$ORIGIN/../lib\" -s -w -X main.version=${BUILD_ID}" -o "${BUILD_NAME}" cmd/photoprism/photoprism.go)
fi

# build binary
echo "=> compiling \"$BUILD_NAME\" with \"${GO_VER^}\""
echo "=> ${BUILD_CMD[*]}"
"${BUILD_CMD[@]}"

# show size
/usr/bin/du -h "${BUILD_NAME}"

echo "Done."