#!/usr/bin/env bash

# This installs NodeJS, NPM and TestCafe on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-nodejs.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

set -e

. /etc/os-release

if [[ $VERSION_CODENAME == "lunar" ]]; then
  echo "Installing NodeJS and NPM distribution packages..."
  apt-get update && apt-get -qq install nodejs npm
else
  SETUP_URL="https://deb.nodesource.com/setup_18.x"

  echo "Fetching packages from \"$SETUP_URL\"..."
  wget --inet4-only -c -qO- $SETUP_URL | bash -

  echo "Installing NodeJS, NPM, and TestCafe..."
  apt-get update && apt-get -qq install nodejs
fi

npm install --unsafe-perm=true --allow-root -g npm testcafe
npm config set cache ~/.cache/npm
npm update --unsafe-perm=true --allow-root -g

echo "Done."