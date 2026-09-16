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
export UA_LOG_LEVEL="info"
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
apt-get -qq install docker-ce docker-ce-cli docker-ce-rootless-extras containerd.io docker-compose-plugin libltdl7 pigz

echo "Adding user..."

# create user
useradd -o -m -U -u 1000 -G docker -d /opt/photoprism photoprism || echo "User 'photoprism' already exists. Proceeding."
mkdir -p /opt/photoprism/originals /opt/photoprism/import /opt/photoprism/storage /opt/photoprism/backup \
      /opt/photoprism/database /opt/photoprism/traefik /opt/photoprism/certs

echo "Generating certificates..."

# download ssl config
curl -fsSL https://dl.photoprism.app/cloud/digitalocean/certs/ca.conf > /opt/photoprism/certs/ca.conf
curl -fsSL https://dl.photoprism.app/cloud/digitalocean/certs/cert.conf > /opt/photoprism/certs/cert.conf
curl -fsSL https://dl.photoprism.app/cloud/digitalocean/certs/config.yml > /opt/photoprism/certs/config.yml
curl -fsSL https://dl.photoprism.app/cloud/digitalocean/certs/openssl.conf > /opt/photoprism/certs/openssl.conf

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

# valid_ip_address reports whether its argument is a plain IPv4 address.
#
# The value is substituted into a YAML document and into the site URL, so it is restricted to the
# characters an address is made of. IPv6 is not accepted: both sources below return IPv4, and an
# IPv6 literal would need bracketing to be a valid URL host, which is a change to make together
# with a source that can return one.
valid_ip_address() {
  local value="$1"

  case "${value}" in
    *[!0-9.]* | "" ) return 1 ;;
  esac

  [[ ${value} =~ ^([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})$ ]] || return 1

  local octet
  for octet in "${BASH_REMATCH[@]:1:4}"; do
    ((10#${octet} <= 255)) || return 1
  done

  return 0
}

# public_ip_address prints the droplet's public address.
#
# The link-local metadata service is authoritative and is reachable only from inside the droplet,
# so it is asked first; the public lookup is a fallback and uses HTTPS, because the value ends up
# in the generated configuration and in the URL the operator is told to trust.
public_ip_address() {
  local address

  # --noproxy, because a proxy variable in the boot environment would otherwise send this to
  # whatever the proxy is, and the answer is only authoritative when it comes from the link-local
  # service itself. --proto and --max-filesize keep the response to the shape of an address.
  address=$(curl -fsS --max-time 5 --noproxy '*' --proto '=http' --max-filesize 64 \
    http://169.254.169.254/metadata/v1/interfaces/public/0/ipv4/address 2>/dev/null || true)

  if valid_ip_address "${address}"; then
    echo "${address}"
    return 0
  fi

  address=$(curl -fsS --max-time 10 --proto '=https' --max-filesize 64 https://api.ipify.org 2>/dev/null || true)

  if valid_ip_address "${address}"; then
    echo "${address}"
    return 0
  fi

  return 1
}

# generate independent random secrets
#
# Each account gets its own, so they can be rotated and shared independently. Hex avoids every
# character that would need quoting in YAML, a shell command or a database connection string.
ADMIN_PASSWORD=$(openssl rand -hex 12)
DATABASE_PASSWORD=$(openssl rand -hex 24)
DATABASE_ROOT_PASSWORD=$(openssl rand -hex 24)

if [[ -z ${ADMIN_PASSWORD} || -z ${DATABASE_PASSWORD} || -z ${DATABASE_ROOT_PASSWORD} ]]; then
  echo "Failed to generate the initial secrets." 1>&2
  exit 1
fi

install -m 600 /dev/null /root/.initial-password.txt
echo "${ADMIN_PASSWORD}" > /root/.initial-password.txt

# detect public server ip address
if ! PUBLIC_IP=$(public_ip_address); then
  echo "Could not determine a valid public IP address for this server." 1>&2
  exit 1
fi

echo "Downloading configuration..."

# download service config
COMPOSE_CONFIG=$(curl -fsSL https://dl.photoprism.app/cloud/digitalocean/compose.yaml)

# Every placeholder must be present before anything is substituted.
#
# This script is baked into the image while the configuration is fetched at boot, so the two can
# be of different vintages. Checking only for leftovers afterwards would not notice a document
# that never carried a placeholder: its substitution is a no-op, nothing is left over, and the
# service starts with whichever secret the remaining placeholders happened to receive. The two
# artifacts are therefore published together, and a mismatch stops the install.
for placeholder in _public_ip_ _admin_password_ _database_password_ _database_root_password_; do
  if [[ ${COMPOSE_CONFIG} != *"${placeholder}"* ]]; then
    echo "The downloaded configuration is missing \"${placeholder}\" and does not match this setup script." 1>&2
    exit 1
  fi
done

COMPOSE_CONFIG=${COMPOSE_CONFIG//_public_ip_/$PUBLIC_IP}
COMPOSE_CONFIG=${COMPOSE_CONFIG//_admin_password_/$ADMIN_PASSWORD}
COMPOSE_CONFIG=${COMPOSE_CONFIG//_database_password_/$DATABASE_PASSWORD}
COMPOSE_CONFIG=${COMPOSE_CONFIG//_database_root_password_/$DATABASE_ROOT_PASSWORD}

# And none may survive it.
if [[ ${COMPOSE_CONFIG} == *_admin_password_* || ${COMPOSE_CONFIG} == *_database_password_* ||
  ${COMPOSE_CONFIG} == *_database_root_password_* || ${COMPOSE_CONFIG} == *_public_ip_* ]]; then
  echo "The downloaded configuration still holds a placeholder after substitution." 1>&2
  exit 1
fi

# The file holds the database and admin secrets, so it is created private and then written.
install -m 600 /dev/null /opt/photoprism/compose.yaml
echo "${COMPOSE_CONFIG}" > /opt/photoprism/compose.yaml
curl -fsSL https://dl.photoprism.app/cloud/digitalocean/traefik.yaml > /opt/photoprism/traefik.yaml

# change permissions
chown -Rf photoprism:photoprism /opt/photoprism

echo "Cleaning up..."

# clear package cache
apt-get autoclean
apt-get autoremove

# start services using docker-compose
(cd /opt/photoprism && docker compose pull && docker compose stop && docker compose up --remove-orphans -d)

# show the public server URL and where to find the initial admin password
#
# The location rather than the value: this runs under cloud-init, whose captured output is kept
# in a log file of its own and shown in the provider console.
printf "\nServer URL:\n\n  https://%s/\n\nInitial admin password:\n\n  see /root/.initial-password.txt\n\n" "${PUBLIC_IP}"
printf "\nPhotoPrism is now installed and running. For documentation, visit:\n\n  https://docs.photoprism.app/\n\n"