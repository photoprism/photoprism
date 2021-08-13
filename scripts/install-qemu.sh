#!/usr/bin/env bash

apt-get update && apt-get -qq install -y qemu binfmt-support qemu-user-static

docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

sleep 10