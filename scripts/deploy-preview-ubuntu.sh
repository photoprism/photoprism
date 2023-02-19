#!/usr/bin/env bash

# Exit on error.
set -ex

# Use QEMU for multi-arch builds.
scripts/install-qemu.sh

# Build preview image.
make docker-preview-ubuntu

# Wait 2s.
sleep 2

# Build ubuntu-based image.
make docker-demo-ubuntu
