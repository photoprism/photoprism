#!/usr/bin/env bash

set -e

# Install dependencies.
echo "Installing dependencies..."
sudo dnf update -y
sudo dnf install nano make netavark aardvark-dns podman podman-docker -y

# Specify Podman network backend.
cat >/etc/containers/containers.conf <<EOL
[network]
network_backend = "netavark"
EOL

# Install Podman Compose if needed.
if ! command -v "podman-compose" &> /dev/null; then
  sudo dnf install python3 python3-pip python3-devel -y
  sudo -H pip3 install --upgrade pip
  sudo pip3 install python-dotenv
  sudo pip3 install pyyaml
  sudo pip3 install podman-compose
fi

# Start Podman service.
sudo systemctl start podman
sudo systemctl enable podman

# Wait 2 seconds.
sleep 2

# Reset Podman and show version.
podman system reset --force
podman --version

# Download config files.
echo "Downloading Makefile and docker-compose.yml..."
curl -o Makefile https://dl.photoprism.app/podman/Makefile
curl -o docker-compose.yml https://dl.photoprism.app/podman/docker-compose.yml

# Create storage folders.
echo "Creating storage folders..."
mkdir -p import mariadb originals storage
sudo chown 1000:1000 import mariadb originals storage
sudo chmod u+rwx,g+rwx import mariadb originals storage

# Show further instructions.
echo ""
echo "Done! You can now customize your settings in the downloaded docker-compose.yml file:"
echo ">> nano docker-compose.yml"
echo "When you are done with the configuration, run 'make' to download and start PhotoPrism."
echo "After waiting a few moments, you should be able to open the UI in a web browser by navigating to:"
echo ">> http://localhost:2342/ (or the configured site URL if you have changed it)"