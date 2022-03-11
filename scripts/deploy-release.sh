#!/usr/bin/env bash

# exit on error
set -e

# install QEMU for multi-arch builds
scripts/install-qemu.sh

# run tests
scripts/test.sh

# build release images
make docker-release
