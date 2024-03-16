#!/usr/bin/env bash

sudo apt-get update && sudo apt-get -qq install -y binfmt-support qemu-kvm qemu-system \
  qemu-user qemu-user-binfmt qemu-utils qemu-efi-arm qemu-efi-aarch64

docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

sleep 10
