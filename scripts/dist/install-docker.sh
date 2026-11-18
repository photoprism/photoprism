#!/usr/bin/env bash

# Installs Docker on Ubuntu Linux
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-docker.sh)

echo "Installing Docker..."
set -e

# Install distribution packages.
sudo apt-get update
sudo apt-get -qq install ca-certificates curl gnupg lsb-release

# Download the repository signing key and replace an existing keyring only once the
# download is a valid key, e.g. when the script is run again after a release upgrade.
KEY_DIR=$(mktemp -d)
trap 'rm -rf "$KEY_DIR" "${BACKUP_DIR:-}"' EXIT
mkdir -m 700 "$KEY_DIR/gnupg"
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o "$KEY_DIR/docker.asc"
if ! GNUPGHOME="$KEY_DIR/gnupg" gpg --show-keys "$KEY_DIR/docker.asc" > /dev/null 2>&1 ||
  ! GNUPGHOME="$KEY_DIR/gnupg" gpg --dearmor < "$KEY_DIR/docker.asc" > "$KEY_DIR/docker.gpg" 2>/dev/null ||
  [ ! -s "$KEY_DIR/docker.gpg" ]; then
  echo "Error: the downloaded Docker signing key is not valid; the installed keyring was kept." >&2
  exit 1
fi
sudo mkdir -p /etc/apt/keyrings
sudo install -m 644 -o root -g root "$KEY_DIR/docker.gpg" /etc/apt/keyrings/docker.gpg

# Determine the Ubuntu release codename. UBUNTU_CODENAME also covers
# derivatives such as Linux Mint; DOCKER_CODENAME overrides both.
# shellcheck source=/dev/null
. /etc/os-release
DOCKER_CODENAME="${DOCKER_CODENAME:-${UBUNTU_CODENAME:-$(lsb_release -cs)}}"

# Keep copies of the current Docker source files so they can be restored if apt
# rejects the new configuration.
SOURCES_DIR=/etc/apt/sources.list.d
BACKUP_DIR=$(mktemp -d)
for f in docker.sources docker.list; do
  if [ -f "$SOURCES_DIR/$f" ]; then sudo cp -p "$SOURCES_DIR/$f" "$BACKUP_DIR/$f"; fi
done

# Add the repository in deb822 format, which release upgrades also convert
# existing entries to, and remove a one-line "docker.list" that would duplicate it.
sudo tee "$SOURCES_DIR/docker.sources" > /dev/null <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: ${DOCKER_CODENAME}
Components: stable
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/docker.gpg
EOF
sudo rm -f "$SOURCES_DIR/docker.list"

# Let apt check that its sources still parse, e.g. when another file configures
# the Docker repository with a different key, and restore the previous files if not.
if ! APT_ERRORS=$(sudo apt-cache policy 2>&1 >/dev/null); then
  sudo rm -f "$SOURCES_DIR/docker.sources"
  for f in docker.sources docker.list; do
    if [ -f "$BACKUP_DIR/$f" ]; then sudo cp -p "$BACKUP_DIR/$f" "$SOURCES_DIR/$f"; fi
  done
  rm -rf "$BACKUP_DIR"
  echo "$APT_ERRORS" >&2
  echo "Error: apt rejected the Docker repository configuration; the previous source files were restored." >&2
  exit 1
fi
rm -rf "$BACKUP_DIR"

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

# Keep locally modified config files such as /etc/containerd/config.toml when
# upgrading, so that dpkg never stops at a conffile prompt without a terminal.
sudo env DEBIAN_FRONTEND=noninteractive apt-get -qq \
  -o Dpkg::Options::=--force-confdef -o Dpkg::Options::=--force-confold \
  install docker-ce docker-ce-cli docker-ce-rootless-extras containerd.io docker-buildx-plugin docker-compose-plugin $EXTRA_PKGS libltdl7 pigz

# Add docker-compose alias for Compose Plugin.
if [ ! -f "/bin/docker-compose" ]; then
  echo 'docker compose "$@"' | sudo tee /bin/docker-compose
  sudo chmod +x /bin/docker-compose
fi

# Verify installation works.
sudo docker run hello-world