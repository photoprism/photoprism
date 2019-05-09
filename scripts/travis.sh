#!/usr/bin/env bash

if [[ -z $TRAVIS_BRANCH ]]; then
    echo "TRAVIS_BRANCH must be set" 1>&2
    exit 1
fi

if [[ $TRAVIS_BRANCH == "develop" ]]; then
    docker-compose -f docker-compose.travis.yml exec photoprism make all test-js test-codecov install migrate;
else
    docker-compose -f docker-compose.travis.yml exec photoprism make all test install migrate;
fi
