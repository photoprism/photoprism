#!/usr/bin/env bash

# PhotoPrism Cloud Init Script
# ============================
#
# Put this script in
#
#   /var/lib/cloud/scripts/per-instance
#
# so that it runs once when the server is booting for the first time:
#
#   cd /var/lib/cloud/scripts/per-instance
#   curl -fsSL https://dl.photoprism.app/cloud/digitalocean/install_photoprism.sh > install_photoprism.sh
#   chmod 700 install_photoprism.sh
#
# To create a reusable image for DigitalOcean:
#
#   packer init digitalocean.pkr.hcl
#   packer build digitalocean.pkr.hcl
#
# Download packer from https://developer.hashicorp.com/packer/install if not installed yet.
#
# Enjoy!

set -eu

SETUP_URL="https://dl.photoprism.app/cloud/digitalocean/setup.sh"
SETUP_SCRIPT=$(mktemp)

# shellcheck disable=SC2064
trap "rm -f '$SETUP_SCRIPT'" EXIT

# Download the setup script to a file first, so that an incomplete or missing
# download aborts with a visible error instead of silently doing nothing.
if ! curl -fsSL "$SETUP_URL" -o "$SETUP_SCRIPT"; then
  echo "Failed to download $SETUP_URL" 1>&2
  exit 1
fi

if [[ ! -s $SETUP_SCRIPT ]]; then
  echo "Downloaded an empty setup script from $SETUP_URL" 1>&2
  exit 1
fi

bash "$SETUP_SCRIPT"
