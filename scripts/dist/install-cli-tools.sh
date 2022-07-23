#!/usr/bin/env bash

# Installs CLI Tools on Linux
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-cli-tools.sh)

# abort if not executed as root
if [[ $(id -u) != "0" ]]; then
  echo "Usage: run ${0##*/} as root" 1>&2
  exit 1
fi

set -eux;

if ! command -v go &> /dev/null
then
    echo "Go must be installed."
    exit 1
fi

echo "Installing CLI Tools..."

GOBIN="/usr/local/bin" go install github.com/muesli/duf@latest
ln -sf /usr/local/bin/duf /usr/local/bin/df