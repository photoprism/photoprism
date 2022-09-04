#!/usr/bin/env bash

# Exit on error.
set -ex

# Use QEMU for multi-arch builds.
scripts/install-qemu.sh

# Run test suite.
scripts/test.sh

# Build preview image.
make docker-preview

# Wait 2s.
sleep 2

# Build demo image.
make docker-demo
