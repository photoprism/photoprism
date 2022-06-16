#!/usr/bin/env bash

# exit on error
set -e

# install QEMU for multi-arch builds
scripts/install-qemu.sh

# build preview image
make docker-preview-ubuntu

# wait 2s
sleep 2

# build demo image
make docker-demo-ubuntu
