#!/usr/bin/env bash

# Installs Docker on Ubuntu Linux
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-docker.sh)

echo "Installing Docker..."
set -e

# Install distribution packages.
sudo apt-get update
sudo apt-get -qq install ca-certificates curl gnupg lsb-release

# Install key.
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker with Compose Plugin.
sudo apt-get update
sudo apt-get -qq install docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Verify installation works.
sudo docker run hello-world