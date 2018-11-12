#!/usr/bin/env bash

if [[ -z $1 ]] || [[ -z $2 ]]; then
    echo "Please provide a container image name and version" 1>&2
    exit 1
else
    echo "Building 'photoprism/$1:$2'...";
    docker build -t photoprism/$1:latest -t photoprism/$1:$2 -f docker/$1/Dockerfile .
    echo "Done"
fi