#!/usr/bin/env bash

# see https://docs.docker.com/develop/develop-images/build_enhancements/#to-enable-buildkit-builds
export DOCKER_BUILDKIT=1

if [[ -z $1 ]] || [[ -z $2 ]]; then
    echo "Please provide a container image name and architecture string (eg. linux/amd64,linux/arm64,linux/arm/v7)" 1>&2
    exit 1
elif [[ $1 ]] && [[ $2 ]] && [[ -z $3 ]]; then
    DOCKER_TAG=$(date -u +%Y%m%d)
    echo "Building 'photoprism/$1:preview'...";
    docker buildx create --name multibuilder --use
    docker buildx build \
      --platform $2 \
      --no-cache \
      --build-arg BUILD_TAG=$DOCKER_TAG \
      -f docker/$1/multiarch/Dockerfile \
      -t photoprism/$1:preview \
      --load .
    docker buildx rm multibuilder
    echo "Done"
else
    echo "Building 'photoprism/$1:$3'...";
    docker buildx create --name multibuilder --use
    docker buildx build \
      --platform $2 \
      --no-cache \
      --build-arg BUILD_TAG=$3 \
      -f docker/$1/multiarch/Dockerfile \
      -t photoprism/$1:latest \
      -t photoprism/$1:$3 \
      --push .
    docker buildx rm multibuilder
    echo "Done"
fi
