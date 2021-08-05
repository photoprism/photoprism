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
#   apt update && apt dist-upgrade && apt autoclean && apt autoremove
#   cd /var/lib/cloud/scripts/per-instance
#   curl -fsSL https://dl.photoprism.org/docker/cloud/install_photoprism.sh > install_photoprism.sh
#   curl -fsSL https://dl.photoprism.org/docker/cloud/enable_firewall.sh > enable_firewall.sh
#   chmod 700 install_photoprism.sh enable_firewall.sh
#   ./enable_firewall.sh
#
# Installing the ufw firewall as shown above is optional but recommended.
#
# When building a reusable image for DigitalOcean, you should run the following commands next:
#
#   bash <(curl -s https://raw.githubusercontent.com/digitalocean/marketplace-partners/master/scripts/90-cleanup.sh)
#   rm -f /var/log/kern.log /var/log/ufw.log
#   bash <(curl -s https://raw.githubusercontent.com/digitalocean/marketplace-partners/master/scripts/99-img-check.sh)
#   shutdown -h now
#
# Enjoy!

bash <(curl -s https://dl.photoprism.org/docker/cloud/setup.sh)
