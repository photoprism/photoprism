#!/usr/bin/env bash

set -e

# Run tests
scripts/test.sh

# Build images
make docker-preview

sleep 2
docker pull photoprism/photoprism:preview

make docker-demo
