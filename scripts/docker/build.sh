#!/usr/bin/env bash

set -e

# see https://docs.docker.com/develop/develop-images/build_enhancements/#to-enable-buildkit-builds
export DOCKER_BUILDKIT=1

if [[ -z $1 ]] || [[ -z $2 ]]; then
    echo "Usage: ${0##*/} [name] [tag] [/subimage]" 1>&2
    exit 1
fi

NUMERIC='^[0-9]+$'
BUILD_DATE=$(/bin/date -u +%y%m%d)

echo "Building image 'photoprism/$1' from docker/${1/-//}$3/Dockerfile...";

if [[ $1 ]] && [[ -z $2 || $2 == "preview" ]]; then
    echo "Build Tags: preview"

    if [[ $4 ]]; then
      echo "Build Params: $4"
    fi

    docker build $4\
      --no-cache \
      --build-arg BUILD_TAG=$BUILD_DATE \
      -t photoprism/$1:preview \
      -f docker/${1/-//}$3/Dockerfile .
elif [[ $2 =~ $NUMERIC ]]; then
    echo "Build Tags: $2, latest"

    if [[ $4 ]]; then
      echo "Build Params: $4"
    fi

    docker build $4\
      --no-cache \
      --build-arg BUILD_TAG=$2 \
      -t photoprism/$1:latest \
      -t photoprism/$1:$2 \
      -f docker/${1/-//}$3/Dockerfile .
elif [[ $2 == *"preview"* || $2 == *"unstable"* || $2 == *"test"* || $2 == *"local"* || $2 == *"develop"* ]]; then
    echo "Build Tags: $2"

    if [[ $4 ]]; then
      echo "Build Params: $4"
    fi

    docker build $4\
      --no-cache \
      --build-arg BUILD_TAG=$BUILD_DATE \
      -t photoprism/$1:$2 \
      -f docker/${1/-//}$3/Dockerfile .
else
    echo "Build Tags: $BUILD_DATE-$2, $2"

    if [[ $4 ]]; then
      echo "Build Params: $4"
    fi

    docker build $4\
      --no-cache \
      --build-arg BUILD_TAG=$BUILD_DATE-$2 \
      -t photoprism/$1:$2 \
      -t photoprism/$1:$BUILD_DATE-$2 \
      -f docker/${1/-//}$3/Dockerfile .
fi

echo "Done."
