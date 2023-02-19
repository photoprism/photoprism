#!/usr/bin/env bash

# Exit on error.
set -ex

# Use QEMU for multi-arch builds.
scripts/install-qemu.sh

# Build preview image.
make docker-preview-debian

# Wait 2s.
sleep 2

# Build debian-based image.
make docker-demo-debian
