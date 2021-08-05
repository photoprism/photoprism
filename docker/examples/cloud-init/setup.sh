#!/usr/bin/env bash

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
echo 'Acquire::Retries "10";' > /etc/apt/apt.conf.d/80retry && \
echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/80recommends && \
echo 'APT::Install-Suggests "false";' > /etc/apt/apt.conf.d/80suggests && \
echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/80forceyes && \
echo 'APT::Get::Fix-Missing "true";' > /etc/apt/apt.conf.d/80fixmissing

# update operating system
apt-get update
apt-get -qq dist-upgrade

# install dependencies
apt-get -qq install -y --no-install-recommends apt-transport-https ca-certificates \
        curl software-properties-common openssl

# install docker if needed
if ! command -v docker &> /dev/null
then
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
  add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu focal stable"
  apt-get update
  apt-cache policy docker-ce
  apt-get -qq install -y docker-ce
  systemctl status docker
fi

# install docker-compose if needed
if ! command -v docker-compose &> /dev/null
then
  apt-get update
  apt-cache policy docker-ce
  apt-get -qq install -y docker-compose
fi

# create user
useradd photoprism -u 1000 -G docker -o -m -d /photoprism || echo "User 'photoprism' already exists. Proceeding."
mkdir -p /photoprism/originals /photoprism/import /photoprism/storage /photoprism/backup \
      /photoprism/database /photoprism/traefik /photoprism/certs

# download ssl config
curl -fsSL https://dl.photoprism.org/docker/cloud-init/certs/ca.conf > /photoprism/certs/ca.conf
curl -fsSL https://dl.photoprism.org/docker/cloud-init/certs/cert.conf > /photoprism/certs/cert.conf
curl -fsSL https://dl.photoprism.org/docker/cloud-init/certs/config.yml > /photoprism/certs/config.yml
curl -fsSL https://dl.photoprism.org/docker/cloud-init/certs/openssl.conf > /photoprism/certs/openssl.conf

# create ca
openssl genrsa -out /photoprism/certs/ca.key 4096
openssl req -x509 -new -nodes -key /photoprism/certs/ca.key -sha256 -days 365 \
        -out /photoprism/certs/ca.pem -config /photoprism/certs/ca.conf
openssl x509 -outform der -in /photoprism/certs/ca.pem -out /photoprism/certs/ca.crt

# create certs
openssl genrsa -out /photoprism/certs/cert.key 4096
openssl req -new -config /photoprism/certs/openssl.conf -key /photoprism/certs/cert.key \
        -out /photoprism/certs/cert.csr
openssl x509 -req -in /photoprism/certs/cert.csr -CA /photoprism/certs/ca.pem \
        -CAkey /photoprism/certs/ca.key -CAcreateserial \
        -out /photoprism/certs/cert.crt -days 365 -sha256 -extfile /photoprism/certs/cert.conf
openssl pkcs12 -export -in /photoprism/certs/cert.crt -inkey /photoprism/certs/cert.key \
        -out /photoprism/certs/cert.pfx -passout pass:

# generate random password
PASSWORD_PLACEHOLDER="_admin_password_"
ADMIN_PASSWORD=$(gpg --gen-random --armor 2 6)
echo "${ADMIN_PASSWORD}" > /root/.initial-password.txt
chmod 600 /root/.initial-password.txt

# detect public server ip address
PUBLIC_IP=$(curl -sfSL ifconfig.me)

# download service config
COMPOSE_CONFIG=$(curl -fsSL https://dl.photoprism.org/docker/cloud-init/docker-compose.yml)
COMPOSE_CONFIG=${COMPOSE_CONFIG//_public_ip_/$PUBLIC_IP}
COMPOSE_CONFIG=${COMPOSE_CONFIG//$PASSWORD_PLACEHOLDER/$ADMIN_PASSWORD}
echo "${COMPOSE_CONFIG}" > /photoprism/docker-compose.yml
curl -fsSL https://dl.photoprism.org/docker/cloud-init/jobs.ini > /photoprism/jobs.ini
curl -fsSL https://dl.photoprism.org/docker/cloud-init/traefik.yaml > /photoprism/traefik.yaml

# change permissions
chown -Rf photoprism:photoprism /photoprism

# start services using docker-compose
(cd /photoprism && docker-compose up -d)

# show public server URL and initial admin password
printf "\nServer URL:\n\n  https://%s/\n\nInitial admin password:\n\n  %s\n\n" "${PUBLIC_IP}" "${ADMIN_PASSWORD}"