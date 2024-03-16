#!/usr/bin/env bash

# This installs a simple firewall on Ubuntu Linux that only allows incoming http, https and ssh connections.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-firewall.sh)

# Install ufw package if needed:
sudo apt-get update
sudo apt-get -qq install --no-install-recommends ufw

# Basic ufw firewall setup allowing ssh, http, and https:
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow http
sudo ufw allow 403
sudo ufw allow https
sudo ufw logging off
sudo rm -f /var/log/ufw.log
sudo ufw --force enable
