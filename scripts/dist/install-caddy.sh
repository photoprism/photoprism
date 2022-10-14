#!/usr/bin/env bash

# Installs Caddy on Linux
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-caddy.sh)

set -e

echo "Installing Caddy..."

sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy

echo "Done."
