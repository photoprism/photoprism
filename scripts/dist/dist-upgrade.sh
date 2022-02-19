#!/usr/bin/env bash

# abort if the user is not root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run dist-upgrade.sh as root" 1>&2
  exit 1
fi

# fail on errors
set -eu

# disable user interactions
export DEBIAN_FRONTEND="noninteractive"
export TMPDIR="/tmp"

apt-get -y update
apt-get -y dist-upgrade
apt-get -y autoremove

echo "Done."