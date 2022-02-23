#!/usr/bin/env bash

# abort if the user is not root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run install-chrome.sh as root" 1>&2
  exit 1
fi

SYSTEM_ARCH=$("$(dirname "$0")/arch.sh")
INSTALL_ARCH=${2:-$SYSTEM_ARCH}

if [[ $INSTALL_ARCH != "amd64" ]]; then
  echo "Google Chrome (stable) is only available for AMD64."
  exit
fi

. /etc/os-release

echo "Installing Google Chrome (stable) on ${ID} for ${INSTALL_ARCH^^}..."

set -e

wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list'
apt-get update
apt-get -qq install google-chrome-stable

echo "Done."