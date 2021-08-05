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
#   curl -fsSL https://dl.photoprism.org/docker/cloud/install_photoprism.sh > install_photoprism.sh
#   chmod 700 install_photoprism.sh
#
# Enjoy!

bash <(curl -s https://dl.photoprism.org/docker/cloud/setup.sh)
