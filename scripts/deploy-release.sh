#!/usr/bin/env bash

set -e

# Run tests
scripts/test.sh

# Build images
make docker-release
