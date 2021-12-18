#!/usr/bin/env bash

set -e

# see https://docs.docker.com/develop/develop-images/build_enhancements/#to-enable-buildkit-builds
export DOCKER_BUILDKIT=1

NUMERIC='^[0-9]+$'
GOPROXY=${GOPROXY:-'https://proxy.golang.org,direct'}

if [[ -z $1 ]] && [[ -z $2 ]]; then
    echo "Please provide a container image name and version" 1>&2
    exit 1
elif [[ $1 ]] && [[ -z $2 ]]; then
    echo "Building 'photoprism/$1:preview'...";
    DOCKER_TAG=$(date -u +%Y%m%d)
    docker build \
      --no-cache \
      --pull \
      --build-arg BUILD_TAG=$DOCKER_TAG \
      --build-arg GOPROXY \
      --build-arg GODEBUG \
      -t photoprism/$1:preview \
      -f docker/${1/-//}/Dockerfile .
    echo "Done"
elif [[ $2 =~ $NUMERIC ]]; then
    echo "Building 'photoprism/$1:$2'...";
    docker build \
      --no-cache \
      --pull \
      --build-arg BUILD_TAG=$2 \
      --build-arg GOPROXY \
      --build-arg GODEBUG \
      -t photoprism/$1:latest \
      -t photoprism/$1:$2 \
      -f docker/${1/-//}/Dockerfile .
    echo "Done"
else
    echo "Building 'photoprism/$1:$2' in docker/${1/-//}$3/Dockerfile...";
    DOCKER_TAG=$(date -u +%Y%m%d)
    docker build $4\
      --no-cache \
      --pull \
      --build-arg BUILD_TAG=$DOCKER_TAG \
      --build-arg GOPROXY \
      --build-arg GODEBUG \
      -t photoprism/$1:$2 \
      -f docker/${1/-//}$3/Dockerfile .
    echo "Done"
fi
