#!/usr/bin/env bash

# abort if the user is not root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run install-nodejs.sh as root" 1>&2
  exit 1
fi

# install from nodesource.com
curl -sL https://deb.nodesource.com/setup_16.x | bash -
apt-get update && apt-get -qq install nodejs
npm install --unsafe-perm=true --allow-root -g npm
npm config set cache ~/.cache/npm
echo "Installed."