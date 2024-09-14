#!/usr/bin/env bash

# This installs the "duf" and "muffet" admin tools on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-admin-tools.sh)

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

echo "Installing duf, a better df alternative..."
GOBIN="/usr/local/bin" go install github.com/muesli/duf@latest

echo "Installing muffet, a fast website link checker..."
GOBIN="/usr/local/bin" go install github.com/raviqqe/muffet@latest

echo "Installing petname, an RFC1178 implementation to generate pronounceable names..."
GOBIN="/usr/local/bin" go install github.com/dustinkirkland/golang-petname/cmd/petname@latest

# Create a symbolic link for "duf" so that it is used instead of the original "df".
ln -sf /usr/local/bin/duf /usr/local/bin/df