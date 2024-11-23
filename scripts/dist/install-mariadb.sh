#!/usr/bin/env bash

# This installs MariaDB on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-mariadb.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

if [[ -z $1 ]]; then
  PACKAGES="mariadb-client"
else
  PACKAGES=$1
fi

set -e

. /etc/os-release

# Determine the system architecture.
if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

echo "Installing \"$PACKAGES\" distribution packages for ${DESTARCH^^}..."

sudo apt-get update
sudo apt-get -qq install $PACKAGES

echo "Done."