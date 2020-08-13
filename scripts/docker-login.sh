#!/usr/bin/env bash

if [[ -z $DOCKER_PASSWORD ]] || [[ -z $DOCKER_USERNAME ]]; then
    docker login
else
    echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
fi
