#!/usr/bin/env bash

# exit on error
set -ex

# install QEMU for multi-arch builds
scripts/install-qemu.sh

# run tests
scripts/test.sh

# build preview image
make docker-preview

# wait 2s
sleep 2

# build demo image
make docker-demo
