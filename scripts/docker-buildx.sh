#!/usr/bin/env bash

# see https://docs.docker.com/develop/develop-images/build_enhancements/#to-enable-buildkit-builds
export DOCKER_BUILDKIT=1

if [[ -z $1 ]] || [[ -z $2 ]] || [[ -z $3 ]]; then
    echo "Please provide a container image name, version and architecture string (eg. linux/amd64,linux/arm64,linux/arm/v7)" 1>&2
    exit 1
else
    echo "Building 'photoprism/$1:$2'...";
    docker buildx create --name multibuilder --use
    docker buildx build \
      --platform $3 \
      --no-cache \
      --build-arg BUILD_TAG=$2 \
      -f docker/development/multiarch/Dockerfile \
      -t photoprism/$1:latest \
      -t photoprism/$1:$2 \
      --push .
    docker buildx rm multibuilder
    echo "Done"
fi
