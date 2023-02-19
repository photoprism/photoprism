#!/usr/bin/env bash

# Exit on error.
set -ex

# Use QEMU for multi-arch builds.
scripts/install-qemu.sh

# Build develop images.
make docker-develop
