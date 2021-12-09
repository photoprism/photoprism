#!/usr/bin/env bash

# https://docs.docker.com/develop/develop-images/build_enhancements/#to-enable-buildkit-builds
export DOCKER_BUILDKIT=1

if [[ -z $1 ]] || [[ -z $2 ]]; then
    echo "Please provide the image name, and a list of target architectures e.g. linux/amd64,linux/arm64,linux/arm" 1>&2
    exit 1
fi

echo "Removing existing multibuilder..."
docker buildx rm multibuilder 2>/dev/null
sleep 3
scripts/install-qemu.sh || { echo 'failed'; exit 1; }
sleep 3
docker buildx create --name multibuilder --use  || { echo 'failed'; exit 1; }

if [[ $1 ]] && [[ $2 ]] && [[ -z $3 ]]; then
    echo "Building 'photoprism/$1:preview'..."
    DOCKER_TAG=$(date -u +%Y%m%d)
    docker buildx build \
      --platform $2 \
      --pull \
      --no-cache \
      --build-arg BUILD_TAG=$DOCKER_TAG \
      -f docker/$1/Dockerfile \
      -t photoprism/$1:preview \
      --push .
else
    echo "Building 'photoprism/$1:$3'..."
    docker buildx build \
      --platform $2 \
      --pull \
      --no-cache \
      --build-arg BUILD_TAG=$3 \
      -f docker/$1/Dockerfile \
      -t photoprism/$1:latest \
      -t photoprism/$1:$3 \
      --push .
fi

echo "Removing multibuilder..."
docker buildx rm multibuilder

echo "Done"
