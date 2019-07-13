#!/usr/bin/env bash

if [[ -z $DOCKER_PASSWORD ]] || [[ -z $DOCKER_USERNAME ]]; then
    docker login
else
    echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
fi

if [[ -z $1 ]] || [[ -z $2 ]]; then
    echo "Please provide a container image name and version" 1>&2
    exit 1
else
    echo "Pushing 'photoprism/$1:$2' to Docker hub...";
    docker push photoprism/$1:latest
    docker push photoprism/$1:$2
    echo "Done"
fi
