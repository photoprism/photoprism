#!/usr/bin/env bash

# This installs RawTherapee on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-rawtherapee.sh)

PATH="/usr/local/sbin:/usr/sbin:/sbin:/usr/local/bin:/usr/bin:/bin:/scripts:$PATH"

# Abort if not executed as root.
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

if [[ $PHOTOPRISM_ARCH ]]; then
  SYSTEM_ARCH=$PHOTOPRISM_ARCH
else
  SYSTEM_ARCH=$(uname -m)
fi

DESTARCH=${2:-$SYSTEM_ARCH}

set -e

. /etc/os-release

echo "Installing RawTherapee for ${DESTARCH^^}..."

case $DESTARCH in
  amd64 | AMD64 | x86_64 | x86-64)
    if [[ $VERSION_CODENAME == "jammy" ]]; then
      echo "Pinning rawtherapee to Lunar Lobster"
      cat << EOF > /etc/apt/preferences.d/rawtherapee.pref
Package: rawtherapee rawtherapee-data libatkmm-1.6-1v5 libcairomm-1.0-1v5 libglibmm-2.4-1v5 libgtkmm-3.0-1v5 libpangomm-1.4-1v5 liblensfun1 liblensfun-data-v1
Pin: release n=lunar
Pin-Priority: 990
EOF
      echo 'deb http://archive.ubuntu.com/ubuntu/ lunar main' | tee /etc/apt/sources.list.d/rawtherapee.list
      echo 'deb http://archive.ubuntu.com/ubuntu/ lunar universe' | tee -a /etc/apt/sources.list.d/rawtherapee.list
      echo "install-rawtherapee: installing RawTherapee 5.9 ($DESTARCH) for Jammy from Lunar Lobster repository"
      apt-get update
      apt-get -qq install rawtherapee
    fi
    ;;

  arm64 | ARM64 | aarch64)
    if [[ $VERSION_CODENAME == "jammy" ]]; then
      echo "Pinning rawtherapee to Lunar Lobster"
      cat << EOF > /etc/apt/preferences.d/rawtherapee.pref
Package: rawtherapee rawtherapee-data libatkmm-1.6-1v5 libcairomm-1.0-1v5 libglibmm-2.4-1v5 libgtkmm-3.0-1v5 libpangomm-1.4-1v5 liblensfun1 liblensfun-data-v1
Pin: release n=lunar
Pin-Priority: 990
EOF
      echo 'deb http://ports.ubuntu.com/ubuntu-ports/ lunar main' | tee /etc/apt/sources.list.d/rawtherapee.list
      echo 'deb http://ports.ubuntu.com/ubuntu-ports/ lunar universe' | tee -a /etc/apt/sources.list.d/rawtherapee.list
      echo "install-rawtherapee: installing RawTherapee 5.9 ($DESTARCH) for Jammy from Lunar Lobster repository"
      apt-get update
      apt-get -qq install rawtherapee
    fi
    ;;

  *)
    echo "Unsupported Machine Architecture: \"$BUILD_ARCH\"" 1>&2
    exit 0
    ;;
esac

echo "Done."
