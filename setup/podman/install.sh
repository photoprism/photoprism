#!/usr/bin/env bash

set -e

# Install dependencies.
# TODO: Tested successfully on AlmaLinux 8, but requires changes for RHEL8.
echo "Installing dependencies..."
sudo dnf upgrade -y
sudo dnf install nano make netavark aardvark-dns podman podman-docker -y

# Install Podman Compose if needed.
if ! command -v "podman-compose" &> /dev/null; then
  sudo dnf install epel-release -y
  sudo dnf install podman-compose -y
fi

# Start Podman service.
sudo systemctl start podman
sudo systemctl enable podman

# Wait 1 second.
sleep 1

# Show Podman version.
podman --version

# Download config files.
echo "Downloading Makefile and docker-compose.yml..."
curl -o Makefile https://dl.photoprism.app/podman/Makefile
curl -o docker-compose.yml https://dl.photoprism.app/podman/docker-compose.yml

# Create storage folders.
echo "Creating storage folders..."
mkdir -p import database originals storage
sudo chown 1000:1000 import database originals storage
sudo chmod u+rwx,g+rwx import database originals storage

# Show further instructions.
echo ""
echo "Done! You can now customize your settings in the downloaded docker-compose.yml file:"
echo ">> nano docker-compose.yml"
echo "When you are done with the configuration, run 'make' to download and start PhotoPrism."
echo "After waiting a few moments, you should be able to open the UI in a web browser by navigating to:"
echo ">> http://localhost:2342/ (or the configured site URL if you have changed it)"