#!/usr/bin/env bash

# abort if the user is not root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run install-mariadb-client.sh as root" 1>&2
  exit 1
fi

curl -Ls https://downloads.mariadb.com/MariaDB/mariadb_repo_setup | bash -s -- --mariadb-server-version="mariadb-10.6"
apt-get update
apt-get -qq install mariadb-client