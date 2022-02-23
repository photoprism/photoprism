#!/usr/bin/env bash

# abort if the user is not root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run install-nodejs.sh as root" 1>&2
  exit 1
fi

set -e

SETUP_URL="https://deb.nodesource.com/setup_16.x"

echo "Installing NodeJS and NPM from \"$SETUP_URL\"..."

curl -sL $SETUP_URL | bash -
apt-get update && apt-get -qq install nodejs
npm install --unsafe-perm=true --allow-root -g npm testcafe
npm config set cache ~/.cache/npm

echo "Done."