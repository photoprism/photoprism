#!/bin/bash

DEFAULT_TF_VERSION=2.18.0

if [[ "$#" -ge 1 ]]; then
    TF_VERSION=$1
elif [[ -z "$TF_VERSION" ]]; then
    TF_VERSION=$DEFAULT_TF_VERSION
fi

SHA2_VERSION=$(curl -L https://raw.githubusercontent.com/tensorflow/tensorflow/refs/tags/v${TF_VERSION}/tensorflow/tools/toolchains/cross_compile/config/BUILD | \
    grep container-image | awk -F'@' '{ print $2 }' | awk -F':' '{ print $2 }' | tr -d '",') 

docker build --build-arg BUILDER_SHA2=$SHA2_VERSION --build-arg TF_VERSION=$TF_VERSION -t photoprism/tensorflow:$TF_VERSION-cross .
