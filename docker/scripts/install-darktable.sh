#!/usr/bin/env bash

# abort if the user is not root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run install-darktable.sh as root" 1>&2
  exit 1
fi

set -e

. /etc/os-release

if [[ -z $1 ]]; then
    echo "Usage: install-darktable.sh [amd64|arm64|arm]" 1>&2
    exit 1
elif [[ $1 == "amd64" ]]; then
  if [[ $VERSION_CODENAME == "bullseye" ]]; then
    echo 'deb http://download.opensuse.org/repositories/graphics:/darktable/Debian_11/ /' | tee /etc/apt/sources.list.d/graphics:darktable.list
    curl -fsSL https://download.opensuse.org/repositories/graphics:darktable/Debian_11/Release.key | gpg --dearmor | tee /etc/apt/trusted.gpg.d/graphics_darktable.gpg > /dev/null
    apt-get update
    apt-get -qq install darktable
  elif [[ $VERSION_CODENAME == "buster" ]]; then
    echo 'deb http://download.opensuse.org/repositories/graphics:/darktable/Debian_10/ /' | tee /etc/apt/sources.list.d/graphics:darktable.list
    curl -fsSL https://download.opensuse.org/repositories/graphics:darktable/Debian_10/Release.key | gpg --dearmor | tee /etc/apt/trusted.gpg.d/graphics_darktable.gpg > /dev/null
    apt-get update
    apt-get -qq install darktable
  else
    echo "install-darktable: installing standard amd64 (Intel 64-bit) package"
    apt-get -qq install darktable
  fi
  echo "install-darktable: done"
elif [[ $1 == "arm64" ]]; then
  if [[ $VERSION_CODENAME == "bullseye" ]]; then
    apt-get update
    apt-get -qq install -t bullseye-backports darktable
  elif [[ $VERSION_CODENAME == "buster" ]]; then
    apt-get update
    apt-get -qq install -t buster-backports darktable
  else
    echo "install-darktable: installing standard amd64 (ARM 64-bit) package"
    apt-get -qq install darktable
  fi
  echo "install-darktable: done"
else
  echo "install-darktable: unsupported architecture $1"
fi
