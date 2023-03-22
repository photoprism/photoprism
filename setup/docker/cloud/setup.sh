#!/usr/bin/env bash

# PhotoPrism Cloud Setup Script
# =============================

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
apt dist-upgrade 2>/dev/null

# install dependencies
apt-get -qq install --no-install-recommends apt-utils apt-transport-https ca-certificates \
        curl software-properties-common openssl gnupg lsb-release

echo "Installing Docker..."

# add docker repository key
mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

# install docker incl compose plugin
apt-get update
apt-get -qq install docker-ce docker-ce-cli docker-ce-rootless-extras containerd.io docker-compose-plugin cgroupfs-mount libltdl7 pigz

echo "Adding user..."

# create user
useradd -o -m -U -u 1000 -G docker -d /opt/photoprism photoprism || echo "User 'photoprism' already exists. Proceeding."
mkdir -p /opt/photoprism/originals /opt/photoprism/import /opt/photoprism/storage /opt/photoprism/backup \
      /opt/photoprism/database /opt/photoprism/traefik /opt/photoprism/certs

echo "Generating certificates..."

# download ssl config
curl -fsSL https://dl.photoprism.app/docker/cloud/certs/ca.conf > /opt/photoprism/certs/ca.conf
curl -fsSL https://dl.photoprism.app/docker/cloud/certs/cert.conf > /opt/photoprism/certs/cert.conf
curl -fsSL https://dl.photoprism.app/docker/cloud/certs/config.yml > /opt/photoprism/certs/config.yml
curl -fsSL https://dl.photoprism.app/docker/cloud/certs/openssl.conf > /opt/photoprism/certs/openssl.conf

# create ca
openssl genrsa -out /opt/photoprism/certs/ca.key 4096
openssl req -x509 -new -nodes -key /opt/photoprism/certs/ca.key -sha256 -days 365 \
        -out /opt/photoprism/certs/ca.pem -config /opt/photoprism/certs/ca.conf
openssl x509 -outform der -in /opt/photoprism/certs/ca.pem -out /opt/photoprism/certs/ca.crt

# create certs
openssl genrsa -out /opt/photoprism/certs/cert.key 4096
openssl req -new -config /opt/photoprism/certs/openssl.conf -key /opt/photoprism/certs/cert.key \
        -out /opt/photoprism/certs/cert.csr
openssl x509 -req -in /opt/photoprism/certs/cert.csr -CA /opt/photoprism/certs/ca.pem \
        -CAkey /opt/photoprism/certs/ca.key -CAcreateserial \
        -out /opt/photoprism/certs/cert.crt -days 365 -sha256 -extfile /opt/photoprism/certs/cert.conf
openssl pkcs12 -export -in /opt/photoprism/certs/cert.crt -inkey /opt/photoprism/certs/cert.key \
        -out /opt/photoprism/certs/cert.pfx -passout pass:

# generate random password
PASSWORD_PLACEHOLDER="_admin_password_"
ADMIN_PASSWORD=$(gpg --gen-random --armor 2 6)
echo "${ADMIN_PASSWORD}" > /root/.initial-password.txt
chmod 600 /root/.initial-password.txt

# detect public server ip address
PUBLIC_IP=$(curl -sfSL ifconfig.me)

echo "Downloading configuration..."

# download service config
COMPOSE_CONFIG=$(curl -fsSL https://dl.photoprism.app/docker/cloud/docker-compose.yml)
COMPOSE_CONFIG=${COMPOSE_CONFIG//_public_ip_/$PUBLIC_IP}
COMPOSE_CONFIG=${COMPOSE_CONFIG//$PASSWORD_PLACEHOLDER/$ADMIN_PASSWORD}
echo "${COMPOSE_CONFIG}" > /opt/photoprism/docker-compose.yml
curl -fsSL https://dl.photoprism.app/docker/cloud/jobs.ini > /opt/photoprism/jobs.ini
curl -fsSL https://dl.photoprism.app/docker/cloud/traefik.yaml > /opt/photoprism/traefik.yaml

# change permissions
chown -Rf photoprism:photoprism /opt/photoprism

echo "Cleaning up..."

# clear package cache
apt-get autoclean
apt-get autoremove

# start services using docker-compose
(cd /opt/photoprism && docker compose pull && docker compose stop && docker compose up --remove-orphans -d)

# show public server URL and initial admin password
printf "\nServer URL:\n\n  https://%s/\n\nInitial admin password:\n\n  %s\n\n" "${PUBLIC_IP}" "${ADMIN_PASSWORD}"