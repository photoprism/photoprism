#!/usr/bin/env bash

# https://docs.docker.com/develop/develop-images/build_enhancements/#to-enable-buildkit-builds
export DOCKER_BUILDKIT=1

if [[ -z $1 ]] || [[ -z $2 ]]; then
    echo "docker/buildx-multi: image name and architectures required (linux/amd64,linux/arm64,linux/arm)" 1>&2
    exit 1
fi

NUMERIC='^[0-9]+$'
GOPROXY=${GOPROXY:-'https://proxy.golang.org,direct'}

# Kill old multibuilder if still alive.
echo "docker/buildx-multi: removing existing multibuilder..."
docker buildx rm multibuilder 2>/dev/null

# Wait 5 seconds.
sleep 5

# Create new multibuilder.
docker buildx create --name multibuilder --use  || { echo 'failed'; exit 1; }

if [[ $1 ]] && [[ $2 ]] && [[ -z $3 ]]; then
    echo "docker/buildx-multi: 'photoprism/$1:preview'..."
    DOCKER_TAG=$(date -u +%Y%m%d)
    docker buildx build \
      --platform $2 \
      --pull \
      --no-cache \
      --build-arg BUILD_TAG=$DOCKER_TAG \
      --build-arg GOPROXY \
      --build-arg GODEBUG \
      -f docker/${1/-//}/Dockerfile \
      -t photoprism/$1:preview \
      --push .
elif [[ $3 =~ $NUMERIC ]]; then
    echo "docker/buildx-multi: 'photoprism/$1:$3'..."
    docker buildx build \
      --platform $2 \
      --pull \
      --no-cache \
      --build-arg BUILD_TAG=$3 \
      --build-arg GOPROXY \
      --build-arg GODEBUG \
      -f docker/${1/-//}/Dockerfile \
      -t photoprism/$1:latest \
      -t photoprism/$1:$3 \
      --push .
elif [[ $4 ]]; then
    echo "docker/buildx-multi: 'photoprism/$1:$3' from docker/${1/-//}$4/Dockerfile..."
    DOCKER_TAG=$(date -u +%Y%m%d)
    docker buildx build \
      --platform $2 \
      --pull \
      --no-cache \
      --build-arg BUILD_TAG=$DOCKER_TAG \
      --build-arg GOPROXY \
      --build-arg GODEBUG \
      -f docker/${1/-//}$4/Dockerfile \
      -t photoprism/$1:$3 \
      --push .
else
    echo "docker/buildx-multi: 'photoprism/$1:$3' from docker/${1/-//}/Dockerfile..."
    DOCKER_TAG=$(date -u +%Y%m%d)
    docker buildx build \
      --platform $2 \
      --pull \
      --no-cache \
      --build-arg BUILD_TAG=$DOCKER_TAG \
      --build-arg GOPROXY \
      --build-arg GODEBUG \
      -f docker/${1/-//}/Dockerfile \
      -t photoprism/$1:$3 \
      --push .
fi

echo "docker/buildx-multi: removing multibuilder..."
docker buildx rm multibuilder

echo "docker/buildx-multi: done"
