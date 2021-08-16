#!/usr/bin/env bash

apt-get update && apt-get -qq install -y qemu binfmt-support qemu-user-static qemu-system-arm qemu-efi-aarch64

docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

sleep 10