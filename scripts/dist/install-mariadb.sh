#!/usr/bin/env bash

# This installs MariaDB on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-mariadb.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

if [[ -z $1 ]]; then
    echo "Usage: ${0##*/} [package names...]" 1>&2
    exit 1
fi

set -e

. /etc/os-release

if [[ $VERSION_CODENAME == "lunar" ]]; then
  echo "Installing MariaDB distribution packages..."
else
  MARIADB_VERSION="10.10"
  MARIADB_URL="https://downloads.mariadb.com/MariaDB/mariadb_repo_setup"

  if [ ! -f "/etc/apt/sources.list.d/mariadb.list" ]; then
    echo "Installing MariaDB $MARIADB_VERSION package sources..."
    curl -Ls $MARIADB_URL | bash  -s -- --mariadb-server-version="mariadb-$MARIADB_VERSION"
  fi
fi

echo "Installing \"$1\"..."

apt-get update
apt-get -qq install $1

echo "Done."