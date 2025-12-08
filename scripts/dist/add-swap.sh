#!/usr/bin/env bash

# Adds a persistent swap file with a default size of 16G if swap space has not yet been configured.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/add-swap.sh)

# Show usage information if first argument is --help, and if not executed as root.
if [[ $(id -u) != "0" ]] || [[ ${1} == "--help" ]]; then
  echo "Adds a persistent swap file with a default size of 16G if swap space has not yet been configured."
  echo "Usage: run \"${0##*/} [size]\" as root" 1>&2
  exit 0
fi

# Check if swap is already configured.
if [[ $(swapon --show) ]]; then
    echo "Swap space has already been configured:"
    swapon --show
    exit 0
fi

set -e

# Add swap as requested, 16G by default.
SWAP_SIZE=${2:-16G}

fallocate -l "${SWAP_SIZE}" /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' | tee -a /etc/fstab

# Check if swap was added successfully.
echo "A persistent /swapfile with a size of ${SWAP_SIZE} was added:"
swapon --show