#!/usr/bin/env bash

# This installs the Nuclei Vulnerability Scanner on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-nuclei.sh)

# Abort if not executed as root..
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

set -eux;

# Is Go installed?
if ! command -v go &> /dev/null
then
    echo "Go must be installed to run this."
    exit 1
fi

# Install Nuclei via go install.
echo "Installing nuclei, a fast and customizable vulnerability scanner..."
GOBIN="/usr/local/bin" go install github.com/projectdiscovery/nuclei/v2/cmd/nuclei@latest
