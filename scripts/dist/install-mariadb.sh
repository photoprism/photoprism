#!/usr/bin/env bash

# abort if the user is not root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run install-mariadb-client.sh as root" 1>&2
  exit 1
fi

if [[ -z $1 ]]; then
    echo "Usage: install-mariadb.sh [package names...]" 1>&2
    exit 1
fi

set -e

SETUP_URL="https://downloads.mariadb.com/MariaDB/mariadb_repo_setup"

if [ ! -f "/etc/apt/sources.list.d/mariadb.list" ]; then
  echo "Adding MariaDB packages sources from \"$SETUP_URL\"..."
  curl -Ls $SETUP_URL | bash -s -- --mariadb-server-version="mariadb-10.6"
fi

echo "Installing \"$1\"..."

apt-get update
apt-get -qq install $1

echo "Done."