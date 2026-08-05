#!/usr/bin/env bash

# Installs the "duf" and "muffet" admin tools on Linux.
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-admin-tools.sh)
#
# This will install additional admin tools for managing Kubernetes deployments on Linux:
# bash <(curl -s https://raw.githubusercontent.com/photoprism/photoprism/develop/scripts/dist/install-k8s-tools.sh)

set -eux;

# Is Go installed?
if ! command -v go &> /dev/null
then
    echo "Go must be installed to run this."
    exit 1
fi

echo "Installing the duf command to check storage usage..."
sudo GOBIN="/usr/local/bin" go install github.com/muesli/duf@latest
sudo ln -sf /usr/local/bin/duf /usr/local/bin/df

echo "Installing muffet, a tool for checking links..."
# The module path must carry the major version: "muffet@latest" still resolves to the
# v1 line (v1.5.7), so only "/v2" installs a current release.
sudo GOBIN="/usr/local/bin" go install github.com/raviqqe/muffet/v2@latest

echo "Installing petname to generate pronounceable names..."
sudo GOBIN="/usr/local/bin" go install github.com/dustinkirkland/golang-petname/cmd/petname@latest
