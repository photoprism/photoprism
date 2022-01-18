#!/usr/bin/env bash

set -e

# Login
scripts/docker/login.sh

# Run tests
docker-compose -f docker-compose.ci.yml pull
docker-compose -f docker-compose.ci.yml build
trap "docker rm -f -v photoprism-test " INT TERM
docker-compose -f docker-compose.ci.yml run --name photoprism-test --rm -T photoprism make all test install migrate
