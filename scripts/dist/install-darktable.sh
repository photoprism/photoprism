#!/bin/bash

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

set -e

SYSTEM_ARCH=$("$(dirname "$0")/arch.sh")
DESTARCH=${2:-$SYSTEM_ARCH}

. /etc/os-release

echo "Installing Darktable for ${DESTARCH^^}..."

if [[ $DESTARCH == "amd64" ]]; then
  if [[ $VERSION_CODENAME == "bullseye" ]]; then
    apt-get update
    apt-get -qq install -t bullseye-backports darktable
  elif [[ $VERSION_CODENAME == "buster" ]]; then
    apt-get update
    apt-get -qq install -t buster-backports darktable
  else
    echo "install-darktable: installing standard amd64 (Intel 64-bit) package"
    apt-get -qq install darktable
  fi
  echo "Done."
elif [[ $DESTARCH == "arm64" ]]; then
  if [[ $VERSION_CODENAME == "bullseye" ]]; then
    apt-get update
    apt-get -qq install -t bullseye-backports darktable
  elif [[ $VERSION_CODENAME == "buster" ]]; then
    apt-get update
    apt-get -qq install -t buster-backports darktable
  else
    echo "install-darktable: installing standard arm64 (ARM 64-bit) package"
    apt-get -qq install darktable
  fi
  echo "Done."
else
  echo "Unsupported Machine Architecture: $DESTARCH"
fi
