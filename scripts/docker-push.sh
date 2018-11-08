#!/usr/bin/env bash

if [ -z "$DOCKER_PASSWORD" ] || [ -z "$DOCKER_USERNAME" ]; then
    echo "DOCKER_PASSWORD and DOCKER_USERNAME not set in your environment";
    exit 1
fi

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin


if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Please provide a container image name and version" 1>&2
    exit 1
else
    echo "Pushing 'photoprism/$1:$2' to Docker hub...";
    docker push photoprism/$1:latest
    docker push photoprism/$1:$2
    echo "Done"
fi