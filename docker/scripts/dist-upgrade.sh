#!/usr/bin/env bash

# check if user is root
if [[ $(id -u) != "0" ]]; then
  echo "failed, please run as root" 1>&2
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
