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

# "cgroupfs-mount" was retired upstream once systemd took over cgroup v2 setup
# and is no longer shipped on Ubuntu 26.04 (resolute) / Debian trixie+. Include
# it only when apt has an installation candidate; "apt-cache show" is not enough
# because a phantom package record can linger with no installable version (seen
# on resolute), which makes the install line abort with "E: Package
# 'cgroupfs-mount' has no installation candidate" before Docker is installed.
EXTRA_PKGS=""
if apt-cache policy cgroupfs-mount 2>/dev/null | grep -q 'Candidate: [^(]'; then
  EXTRA_PKGS="cgroupfs-mount"
fi

sudo apt-get -qq install docker-ce docker-ce-cli docker-ce-rootless-extras containerd.io docker-buildx-plugin docker-compose-plugin $EXTRA_PKGS libltdl7 pigz

# Add docker-compose alias for Compose Plugin.
if [ ! -f "/bin/docker-compose" ]; then
  echo 'docker compose "$@"' | sudo tee /bin/docker-compose
  sudo chmod +x /bin/docker-compose
fi

# Verify installation works.
sudo docker run hello-world