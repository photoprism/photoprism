#!/usr/bin/env bash

# This installs FFmpeg on Ubuntu Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-ffmpeg.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

# Determine the system architecture.
if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

set -e

. /etc/os-release

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64 | arm64 | ARM64 | aarch64)
    if [[ $VERSION_CODENAME == "noble" ]]; then
      # Install FFmpeg 7 from a PPA on Ubuntu 24.04 LTS.
      echo "Installing FFmpeg 7 for ${DESTARCH^^} from ppa:ubuntuhandbook1/ffmpeg7..."
      apt-get update
      apt-get -qq install software-properties-common pkg-config apt-utils
      add-apt-repository -y ppa:ubuntuhandbook1/ffmpeg7
      apt-get update
      apt-get -qq install ffmpeg
      apt-get -qq dist-upgrade
    else
      # Otherwise, install the default FFmpeg package.
      echo "Installing FFmpeg for ${DESTARCH^^}..."
      apt-get update
      apt-get -qq install ffmpeg
    fi
    ;;

  *)
    # Install the default FFmpeg package.
    echo "Installing FFmpeg for ${DESTARCH^^}..."
    apt-get update
    apt-get -qq install ffmpeg
    ;;
esac

echo "Done."
