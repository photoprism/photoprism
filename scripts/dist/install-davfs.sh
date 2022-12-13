#!/usr/bin/env bash

# This installs the DavFS filesystem driver on Linux.

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

echo "Installing WebDAV filesystem driver..."

apt-get update
apt-get -qq install davfs2

echo "Done."