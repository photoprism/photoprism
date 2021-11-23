#!/usr/bin/env bash

set -e

# Login
scripts/docker-login.sh

# Test
docker-compose -f docker-compose.ci.yml pull
docker-compose -f docker-compose.ci.yml stop
docker-compose -f docker-compose.ci.yml up -d --build --force-recreate
docker-compose -f docker-compose.ci.yml exec -T photoprism make all test install migrate
docker-compose -f docker-compose.ci.yml down

# Deploy
scripts/install-qemu.sh
sleep 2
make docker-preview
sleep 2
docker pull photoprism/photoprism:preview
make docker-demo