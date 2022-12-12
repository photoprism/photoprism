#!/usr/bin/env bash

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
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