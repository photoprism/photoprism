#!/usr/bin/env bash

set -e

# Run tests
scripts/test.sh

# Build images
scripts/install-qemu.sh
sleep 2
make docker-preview
sleep 2
docker pull photoprism/photoprism:preview
make docker-demo