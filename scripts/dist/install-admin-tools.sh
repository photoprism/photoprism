#!/usr/bin/env bash

# Installs Admin Tools on Linux
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-admin-tools.sh)

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

set -eux;

if ! command -v go &> /dev/null
then
    echo "Go must be installed to run this."
    exit 1
fi

echo "Installing duf, a better df alternative..."
GOBIN="/usr/local/bin" go install github.com/muesli/duf@latest

echo "Installing muffet, a fast website link checker..."
GOBIN="/usr/local/bin" go install github.com/raviqqe/muffet@latest

echo "Installing nuclei, a fast and customizable vulnerability scanner..."
GOBIN="/usr/local/bin" go install github.com/projectdiscovery/nuclei/v2/cmd/nuclei@latest

ln -sf /usr/local/bin/duf /usr/local/bin/df