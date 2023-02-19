#!/usr/bin/env bash

# Exit on error.
set -ex

# Use QEMU for multi-arch builds.
scripts/install-qemu.sh

# Run test suite.
scripts/test.sh

# Build release image.
make docker-release
