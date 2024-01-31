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

# NodeJS version.
NODE_MAJOR=20

# Create /etc/apt/keyrings/nodesource.gpg
mkdir -p /etc/apt/keyrings
curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg

# Create /etc/apt/sources.list.d/nodesource.list
echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_$NODE_MAJOR.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list

echo "Installing NodeJS and NPM from deb.nodesource.com..."
apt-get update && apt-get -qq install nodejs

echo "Installing TestCafe..."
npm config set cache ~/.cache/npm
npm install --unsafe-perm=true --allow-root -g npm@latest npm-check-updates@latest n@latest testcafe@3.4.0

echo "Done."