#!/usr/bin/env bash

# Installs PostgreSQL packages on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-postgresql.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

if [[ $# -eq 0 ]]; then
  PACKAGES=("postgresql-client")
else
  PACKAGES=("$@")
fi

set -e

# shellcheck source=/dev/null
. /etc/os-release

# Determine target architecture.
if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${BUILD_ARCH:-$SYSTEM_ARCH}

echo "Installing \"${PACKAGES[*]}\" distribution packages for ${DESTARCH^^}..."

sudo apt-get update
sudo apt-get -qy install curl gnupg postgresql-common apt-transport-https lsb-release
sudo sh /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh -y
sudo apt-get -qq install "${PACKAGES[@]}"

echo "Done."
