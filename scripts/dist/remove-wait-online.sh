#!/usr/bin/env bash

# This script disables the wait-online service to prevent the system from waiting for a network connection.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/remove-wait-online.sh)

set -e

echo "Disabling the wait-online service..."

sudo systemctl disable systemd-networkd-wait-online.service && \
  sudo systemctl mask systemd-networkd-wait-online.service

echo "Done."
