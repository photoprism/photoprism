#!/usr/bin/env bash

# exit on error
set -ex

# Login
scripts/docker/login.sh

# Run tests
docker-compose -f docker-compose.ci.yml down --remove-orphans
docker-compose -f docker-compose.ci.yml pull
docker-compose -f docker-compose.ci.yml build --pull
trap "docker-compose -f docker-compose.ci.yml down --remove-orphans" INT TERM
docker-compose -f docker-compose.ci.yml run --rm photoprism make all test install migrate
docker-compose -f docker-compose.ci.yml down --remove-orphans
