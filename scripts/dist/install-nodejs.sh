#!/usr/bin/env bash

# Installs NodeJS, NPM and TestCafe on Linux
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-nodejs.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

set -e

SETUP_URL="https://deb.nodesource.com/setup_18.x"
wget --inet4-only -c -qO- https://deb.nodesource.com/setup_18.x | bash -
echo "Installing NodeJS and NPM from \"$SETUP_URL\"..."

curl -sL $SETUP_URL | bash  -
apt-get update && apt-get -qq install nodejs
npm install --unsafe-perm=true --allow-root -g npm testcafe
npm config set cache ~/.cache/npm
npm update --unsafe-perm=true --allow-root -g

echo "Done."