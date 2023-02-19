#!/usr/bin/env bash

# Login to Docker Hub.
scripts/docker/login.sh

# Define functions.
cleanUp() {
  COMPOSE_PROJECT_NAME=ci docker-compose -f docker-compose.ci.yml down --remove-orphans
}

# Make sure containers are not running and don't keep running.
cleanUp
trap cleanUp INT

# Set up environment and run tests.
ERROR=0
COMPOSE_PROJECT_NAME=ci docker-compose -f docker-compose.ci.yml pull --ignore-pull-failures && \
COMPOSE_PROJECT_NAME=ci docker-compose -f docker-compose.ci.yml build --pull && \
COMPOSE_PROJECT_NAME=ci docker-compose -f docker-compose.ci.yml run --rm photoprism make all test install migrate || \
ERROR=1

# Stop containers.
cleanUp

# Failed?
if [[ $ERROR == "1" ]]; then
  echo "Failed."
  exit 1
fi

echo "Done."
