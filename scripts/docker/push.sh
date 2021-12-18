#!/usr/bin/env bash

set -e

if [[ -z $DOCKER_PASSWORD ]] || [[ -z $DOCKER_USERNAME ]]; then
    docker login
fi

NUMERIC='^[0-9]+$'

if [[ -z $1 ]] && [[ -z $2 ]]; then
    echo "Please provide a container image name and version" 1>&2
    exit 1
elif [[ $1 ]] && [[ -z $2 ]]; then
    echo "Pushing 'photoprism/$1:preview' to Docker hub...";
    docker push photoprism/$1:preview
    echo "Done"
elif [[ $2 =~ $NUMERIC ]]; then
    echo "Pushing 'photoprism/$1:$2' to Docker hub...";
    docker push photoprism/$1:latest
    docker push photoprism/$1:$2
    echo "Done"
else
    echo "Pushing 'photoprism/$1:$2' to Docker hub...";
    docker push photoprism/$1:$2
    echo "Done"
fi
