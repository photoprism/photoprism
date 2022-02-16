#!/usr/bin/env bash

# https://docs.docker.com/develop/develop-images/build_enhancements/#to-enable-buildkit-builds
export DOCKER_BUILDKIT=1

if [[ -z $1 ]] || [[ -z $2 ]]; then
    echo "usage: scripts/docker/buildx.sh [image] linux/[amd64|arm64|arm] [tag] [/subimage]" 1>&2
    exit 1
fi

NUMERIC='^[0-9]+$'
GOPROXY=${GOPROXY:-'https://proxy.golang.org,direct'}
DOCKER_TAG=$(date -u +%Y%m%d)

if [[ $1 ]] && [[ $2 ]] && [[ -z $3 || $3 == "preview" ]]; then
    echo "docker/buildx: building photoprism/$1:preview from docker/${1/-//}$4/Dockerfile..."
    docker buildx build \
      --platform $2 \
      --pull \
      --no-cache \
      --build-arg BUILD_TAG=$DOCKER_TAG \
      --build-arg GOPROXY \
      --build-arg GODEBUG \
      -f docker/${1/-//}$4/Dockerfile \
      -t photoprism/$1:preview \
      --push .
elif [[ $3 =~ $NUMERIC ]]; then
    echo "docker/buildx: building photoprism/$1:$3,$1:latest from docker/${1/-//}$4/Dockerfile..."

    if [[ $5 ]]; then
      echo "extra params: $5"
    fi

    docker buildx build \
      --platform $2 \
      --pull \
      --no-cache \
      --build-arg BUILD_TAG=$3 \
      --build-arg GOPROXY \
      --build-arg GODEBUG \
      -f docker/${1/-//}$4/Dockerfile \
      -t photoprism/$1:latest \
      -t photoprism/$1:$3 $5 \
      --push .
elif [[ $4 ]] && [[ $3 == *"preview"* ]]; then
    echo "docker/buildx: building photoprism/$1:$3 from docker/${1/-//}$4/Dockerfile..."

    if [[ $5 ]]; then
      echo "extra params: $5"
    fi

    docker buildx build \
      --platform $2 \
      --pull \
      --no-cache \
      --build-arg BUILD_TAG=$DOCKER_TAG \
      --build-arg GOPROXY \
      --build-arg GODEBUG \
      -f docker/${1/-//}$4/Dockerfile \
      -t photoprism/$1:$3 $5 \
      --push .
else
    echo "docker/buildx: building photoprism/$1:$3,$1:$DOCKER_TAG-$3 from docker/${1/-//}$4/Dockerfile..."

    if [[ $5 ]]; then
      echo "extra params: $5"
    fi

    docker buildx build \
      --platform $2 \
      --pull \
      --no-cache \
      --build-arg BUILD_TAG=$DOCKER_TAG \
      --build-arg GOPROXY \
      --build-arg GODEBUG \
      -f docker/${1/-//}$4/Dockerfile \
      -t photoprism/$1:$3 \
      -t photoprism/$1:$DOCKER_TAG-$3 $5 \
      --push .
fi

echo "docker/buildx: done"
