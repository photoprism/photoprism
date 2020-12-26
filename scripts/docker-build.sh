#!/usr/bin/env bash

if [[ -z $1 ]] && [[ -z $2 ]]; then
    echo "Please provide a container image name and version" 1>&2
    exit 1
elif [[ $1 ]] && [[ -z $2 ]]; then
    DOCKER_TAG=$(date -u +%Y%m%d)
    echo "Building 'photoprism/$1:preview'...";
    docker build --no-cache --build-arg BUILD_TAG="${DOCKER_TAG}" -t photoprism/$1:preview -f docker/${1/-//}/Dockerfile .
    echo "Done"
else
    echo "Building 'photoprism/$1:$2'...";
    docker build --no-cache --build-arg BUILD_TAG=$2 -t photoprism/$1:latest -t photoprism/$1:$2 -f docker/${1/-//}/Dockerfile .
    echo "Done"
fi
