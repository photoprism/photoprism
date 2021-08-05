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
apt-get -qq install -y --no-install-recommends apt-transport-https ca-certificates curl software-properties-common

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
useradd photoprism -u 1000 -G docker -o -m -d /photoprism
mkdir -p /photoprism/originals /photoprism/import /photoprism/storage /photoprism/backup /photoprism/database

# download service config
curl -fsSL https://dl.photoprism.org/docker/cloud-init/docker-compose.yml> /photoprism/docker-compose.yml
curl -fsSL https://dl.photoprism.org/docker/cloud-init/jobs.ini > /photoprism/jobs.ini
chown -Rf photoprism:photoprism /photoprism

# start services using docker-compose
(cd /photoprism && docker-compose up -d)
