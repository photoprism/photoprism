#!/bin/bash

# abort if not executed as root
if [[ $(/usr/bin/id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

# fail on errors
set -eu

# disable user interactions
export DEBIAN_FRONTEND="noninteractive"
export TMPDIR="/tmp"

/usr/bin/apt-get -y update
/usr/bin/apt-get -y dist-upgrade
/usr/bin/apt-get -y autoremove

echo "Done."