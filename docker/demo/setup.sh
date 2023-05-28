#!/usr/bin/env bash

# PhotoPrism Demo Environment - Setup Script
#
# Usage:
#   bash <(curl -s https://dl.photoprism.app/docker/demo/setup.sh)
#
# Note:
# - demo.yourdomain.com must be replaced with the actual hostname
# - The demo is frequently restarted to remove uploaded content
# - There is no password protection, it is running in public mode


# check if user is root
if [[ $(id -u) != "0" ]]; then
  echo "You need to be root to run this script." 1>&2
  exit 1
fi

# fail on errors
set -eu

# disable user interactions
export DEBIAN_FRONTEND="noninteractive"
export TMPDIR="/tmp"

# add 4 GB of swap if no swap was configured yet
if [[ -z $(swapon --show) ]]; then
  fallocate -l 4G /swapfile
  chmod 600 /swapfile
  mkswap /swapfile
  swapon /swapfile
  swapon --show
  free -h
  echo '/swapfile none swap sw 0 0' | tee -a /etc/fstab
fi

# set apt defaults
echo 'APT::Acquire::Retries "3";' > /etc/apt/apt.conf.d/80retries && \
echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissing && \
echo 'force-confold' > /etc/dpkg/dpkg.cfg.d/force-confold

# update operating system
apt-get update
apt upgrade 2>/dev/null

# install dependencies
apt-get -qq install --no-install-recommends apt-transport-https ca-certificates \
        curl wget make software-properties-common net-tools openssl ufw

# install docker if needed
if ! command -v docker &> /dev/null; then
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/trusted.gpg.d/download.docker.com.gpg
  add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu focal stable"
  apt-get update
  apt-get -qq install docker-ce
fi

# install docker-compose if needed
if ! command -v docker-compose &> /dev/null; then
  apt-get update
  apt-get -qq install docker-compose
fi

# Basic ufw firewall setup allowing ssh, http, and https
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow http
ufw allow https
ufw logging off
rm -f /var/log/ufw.log
ufw --force enable

# create user
useradd -o -m -U -u 1000 -G docker -d /opt/photoprism photoprism || echo "User 'photoprism' already exists. Proceeding."
mkdir -p /opt/photoprism/config /opt/photoprism/traefik

# detect public server ip address
PUBLIC_IP=$(curl -sfSL ifconfig.me)

# download service config
COMPOSE_CONFIG=$(curl -fsSL https://dl.photoprism.app/docker/demo/docker-compose.yml)
COMPOSE_CONFIG=${COMPOSE_CONFIG//_public_ip_/$PUBLIC_IP}
echo "${COMPOSE_CONFIG}" > /opt/photoprism/docker-compose.yml
curl -fsSL https://dl.photoprism.app/docker/demo/jobs.ini > /opt/photoprism/jobs.ini
curl -fsSL https://dl.photoprism.app/docker/demo/traefik.yaml > /opt/photoprism/traefik.yaml
curl -fsSL https://dl.photoprism.app/docker/demo/Makefile > /opt/photoprism/Makefile

# change permissions
chown -Rf photoprism:photoprism /opt/photoprism

# clear package cache
apt-get -y autoclean
apt-get -y autoremove

# start services using docker-compose
(cd /opt/photoprism && make install)
