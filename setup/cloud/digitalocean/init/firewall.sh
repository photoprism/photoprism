#!/usr/bin/env bash

# Install ufw package if needed:
apt-get update && apt-get -y install --no-install-recommends ufw && apt-get -y autoclean && apt-get -y autoremove

# Basic ufw firewall setup allowing ssh, http, and https:
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow http
ufw allow https
# Discard anything logged while the image was being built, so a droplet starts with a log that
# only describes itself.
rm -f /var/log/ufw.log

# Keep a low-volume record of what the firewall blocked from here on, so an operator
# investigating an incident has something to look at. "low" logs blocked packets only,
# rate-limited by ufw.
ufw logging low
ufw --force enable
