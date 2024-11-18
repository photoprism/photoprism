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

if [[ $VERSION_CODENAME == "oracular" || $DESTARCH == "armv7l" || $DESTARCH == "arm" ]]; then
  echo "Installing MariaDB distribution packages for ${DESTARCH^^}..."
else
  MARIADB_VERSION="11.rolling"
  MARIADB_URL="https://downloads.mariadb.com/MariaDB/mariadb_repo_setup"

  if [ ! -f "/etc/apt/sources.list.d/mariadb.list" ]; then
    echo "Installing MariaDB $MARIADB_VERSION package sources for ${DESTARCH^^}..."
    curl -Ls "$MARIADB_URL" | sudo bash -s -- --mariadb-server-version="mariadb-$MARIADB_VERSION"
  fi
fi

echo "Installing \"$PACKAGES\"..."

sudo apt-get update
sudo apt-get -qq install $PACKAGES

echo "Done."