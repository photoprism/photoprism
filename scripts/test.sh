#!/usr/bin/env bash

set -e

# Login
scripts/docker-login.sh

# Run tests
docker-compose -f docker-compose.ci.yml pull
docker-compose -f docker-compose.ci.yml stop
docker-compose -f docker-compose.ci.yml up -d --build --force-recreate
docker-compose -f docker-compose.ci.yml exec -T photoprism make all test install migrate
docker-compose -f docker-compose.ci.yml down
