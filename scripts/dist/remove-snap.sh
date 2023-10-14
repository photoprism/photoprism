#!/usr/bin/env bash

# This script removes and disables Snap on Ubuntu Linux. See <https://www.baeldung.com/linux/snap-remove-disable>.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/remove-snap.sh)

set -e

echo "Disabling Snap..."

sudo systemctl stop snapd && \
  sudo systemctl disable snapd && \
  sudo systemctl mask snapd

echo "Removing Snap..."

sudo apt-get purge snapd -y && \
  sudo apt-mark hold snapd && \
  sudo rm -rf /snap /root/snap /var/snap /var/lib/snapd

echo "Done."
