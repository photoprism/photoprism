#!/bin/bash

PATH="/usr/local/sbin/:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts"

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

SYSTEM_ARCH=$("$(dirname "$0")/arch.sh")
DESTARCH=${2:-$SYSTEM_ARCH}
. /etc/os-release

if [[ $DESTARCH != "amd64" ]]; then
  echo "Google Chrome (stable) is only available for AMD64."
  exit
fi

echo "Installing Google Chrome (stable) on ${ID} for ${DESTARCH^^}..."

set -e

wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list'
apt-get update
apt-get -qq install google-chrome-stable

echo "Done."