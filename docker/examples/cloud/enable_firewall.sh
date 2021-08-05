#!/usr/bin/env bash

# Install ufw package if needed:
apt-get update && apt-get install -y --no-install-recommends ufw && apt-get autoclean && apt-get autoremove

# Basic ufw firewall setup allowing ssh, http, and https:
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow http
ufw allow https
ufw --force enable
