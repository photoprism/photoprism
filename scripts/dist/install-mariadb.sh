#!/usr/bin/env bash

# Installs MariaDB on Linux
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-mariadb.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

if [[ -z $1 ]]; then
    echo "Usage: ${0##*/} [package names...]" 1>&2
    exit 1
fi

set -e

SETUP_URL="https://downloads.mariadb.com/MariaDB/mariadb_repo_setup"

if [ ! -f "/etc/apt/sources.list.d/mariadb.list" ]; then
  echo "Adding MariaDB packages sources from \"$SETUP_URL\"..."
  curl -Ls $SETUP_URL | bash  -s -- --mariadb-server-version="mariadb-10.9"
fi

echo "Installing \"$1\"..."

apt-get update
apt-get -qq install $1

echo "Done."