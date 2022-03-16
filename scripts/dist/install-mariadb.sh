#!/usr/bin/env bash

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
  /usr/bin/curl -Ls $SETUP_URL | /bin/bash  -s -- --mariadb-server-version="mariadb-10.6"
fi

echo "Installing \"$1\"..."

/usr/bin/apt-get update
/usr/bin/apt-get -qq install $1

echo "Done."