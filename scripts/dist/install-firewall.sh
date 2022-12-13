#!/usr/bin/env bash

# Install ufw package if needed:
sudo apt-get update
sudo apt-get -qq install --no-install-recommends ufw

# Basic ufw firewall setup allowing ssh, http, and https:
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow http
sudo ufw allow https
sudo ufw logging off
sudo rm -f /var/log/ufw.log
sudo ufw --force enable
