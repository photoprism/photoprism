#!/usr/bin/env bash

set -e

# Login
scripts/docker/login.sh

# Run tests
docker-compose -f docker-compose.ci.yml down --remove-orphans
docker-compose -f docker-compose.ci.yml pull
docker-compose -f docker-compose.ci.yml build
trap "docker-compose -f docker-compose.ci.yml down --remove-orphans" INT TERM
docker-compose -f docker-compose.ci.yml run --rm -T photoprism make all test install migrate
docker-compose -f docker-compose.ci.yml down --remove-orphans
