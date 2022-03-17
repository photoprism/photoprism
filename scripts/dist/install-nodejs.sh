#!/bin/bash

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

set -e

SETUP_URL="https://deb.nodesource.com/setup_16.x"

echo "Installing NodeJS and NPM from \"$SETUP_URL\"..."

/usr/bin/curl -sL $SETUP_URL | /bin/bash  -
/usr/bin/apt-get update && /usr/bin/apt-get -qq install nodejs
npm install --unsafe-perm=true --allow-root -g npm testcafe
npm config set cache ~/.cache/npm

echo "Done."