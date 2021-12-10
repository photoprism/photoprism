#!/usr/bin/env bash

sudo apt-get update && sudo apt-get -qq install -y qemu binfmt-support qemu-kvm qemu-system \
  qemu-user qemu-user-binfmt qemu-user-static qemu-utils \
  qemu-efi-arm qemu-efi-aarch64 libvirt-daemon-driver-qemu

docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

sleep 10