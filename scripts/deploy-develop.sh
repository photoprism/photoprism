#!/usr/bin/env bash

# exit on error
set -ex

# install QEMU for multi-arch builds
scripts/install-qemu.sh

# build release images
make docker-develop-all
